package scheduler

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/digitalocean/godo"
	pb "github.com/lgpeterson/loadtests/executor/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	executorPrefix = "executor"
	bootSequence   = `#!/usr/bin/env bash

echo "the boot script started" > /tmp/bootscript.log
mkdir -p /opt
curl %q > /opt/executord
chmod +x /opt/executord

cat > /etc/systemd/system/load_executor.service <<EOF
[Unit]
Description=Load executor service

[Service]
ExecStart=/opt/executord -scheduler_addr %q
Restart=always
RestartSec=1

[Install]
WantedBy=multi-user.target
EOF

systemctl enable load_executor.service
systemctl start load_executor.service
`
)

type DB struct {
	cfg          *Config
	cloud        *godo.Client
	lock         sync.Mutex
	waitDroplets map[int]chan<- int
}

func NewDB(cfg *Config, cloud *godo.Client) (*DB, error) {

	// cleanup any executors that are still running, if we crashed
	droplets, _, err := cloud.Droplets.List(&godo.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, droplet := range droplets {
		if !strings.HasPrefix(droplet.Name, executorPrefix) {
			continue
		}
		_, err := cloud.Droplets.Delete(droplet.ID)
		if err != nil {
			return nil, err
		}
	}

	return &DB{cfg: cfg, cloud: cloud, waitDroplets: make(map[int]chan<- int)}, nil
}

func (db *DB) LaunchExecutors(ctx context.Context, count int) (*executors, error) {
	logrus.WithField("count", count).Info("launching executors")
	var (
		wg        sync.WaitGroup
		exec      = new(executors)
		executorc = make(chan *executor, count)
		errc      = make(chan error, count)
	)

	suffix := fmt.Sprintf("%d-%d", rand.Int(), time.Now().UTC().Unix())

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := &godo.DropletCreateRequest{
				Name:              fmt.Sprintf("%s.%s.%d", executorPrefix, suffix, id),
				SSHKeys:           db.cfg.SSHKeyIDs,
				PrivateNetworking: true,
				Region:            db.cfg.DropletRegion,
				Size:              db.cfg.DropletSize,
				UserData: fmt.Sprintf(bootSequence,
					db.cfg.PullExecutorBinaryURL,
					db.cfg.AdvertiseListenAddr,
				),
				Image: godo.DropletCreateImage{Slug: db.cfg.DropletImageSlug},
			}

			db.lock.Lock()
			droplet, _, err := db.cloud.Droplets.Create(req)
			if err != nil {
				defer db.lock.Unlock()
				errc <- err
				return
			}
			portc := make(chan int, 1)
			db.waitDroplets[droplet.ID] = portc
			db.lock.Unlock()
			defer func() {
				db.lock.Lock()
				delete(db.waitDroplets, droplet.ID)
				db.lock.Unlock()
			}()

			select {
			case port := <-portc:
				droplet, _, err := db.cloud.Droplets.Get(droplet.ID)
				if err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"port":       port,
						"droplet.id": droplet.ID,
					}).Error("failed to retrieve details about executor")
					errc <- err
					return
				}

				ip, found := ipv4PublicAddress(droplet)
				if !found {
					errc <- fmt.Errorf("no public IPv4 found on droplet %d", droplet.ID)
					return
				}

				logrus.WithFields(logrus.Fields{
					"port":       port,
					"droplet.id": droplet.ID,
					"droplet.ip": ip,
				}).Info("executor joined")
				executorc <- &executor{
					cloud:   db.cloud,
					droplet: droplet,
					ip:      ip,
					port:    port,
				}
			case <-ctx.Done():
				logrus.WithFields(logrus.Fields{
					"droplet.id": droplet.ID,
				}).Info("timedout waiting for executor")
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(executorc)
	}()

	for {
		select {
		case executor, more := <-executorc:
			if !more {
				logrus.Info("all executors have joined")
				return exec, nil
			}
			exec.executors = append(exec.executors, executor)
		case err := <-errc:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func ipv4PublicAddress(droplet *godo.Droplet) (string, bool) {
	if droplet.Networks == nil {
		return "", false
	}
	for _, network := range droplet.Networks.V4 {
		switch network.Type {
		case "public":
			return network.IPAddress, true
		}
	}
	return "", false
}

func (db *DB) RegisterExecutorUp(dropletID int, port int) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	ll := logrus.WithFields(logrus.Fields{
		"droplet.id": dropletID,
		"port":       port,
	})
	ll.Info("executor is registering")
	wait, ok := db.waitDroplets[dropletID]
	if !ok {
		ll.Warn("unexpected executor attempted to join")
		// _, err := db.cloud.Droplets.Delete(dropletID)
		var err error
		return fmt.Errorf("unexpected droplet %d registered, delete request sent: %v", dropletID, err)
	}
	wait <- port
	return nil
}

type executor struct {
	cloud   *godo.Client
	droplet *godo.Droplet
	ip      string
	port    int
	client  pb.CommanderClient

	// set when there's an ongoing command execution
	cmdClient pb.Commander_ExecuteCommandClient
}

func (e *executor) waitTilAlive(ctx context.Context) error {
	if e.client != nil {
		return nil
	}
	url := fmt.Sprintf("%s:%d", e.ip, e.port)
	ll := logrus.WithFields(logrus.Fields{
		"droplet.id":   e.droplet.ID,
		"droplet.ip":   e.ip,
		"executor.url": url,
	})
	ll.Info("attempting to dial executor service")
	for {
		if e.client != nil {
			ll.Info("service reached")
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		cc, err := grpc.Dial(url, grpc.WithBlock(), grpc.WithTimeout(time.Second), grpc.WithInsecure())
		switch err {
		case grpc.ErrClientConnTimeout:
			ll.Info("timed out...")
			continue
		case nil:
			e.client = pb.NewCommanderClient(cc)
		default:
			ll.WithError(err).Error("couldn't dial executor service")
			return err
		}
	}
}

type executors struct {
	cloud     *godo.Client
	executors []*executor
}

func (e *executors) killall() error {
	logrus.WithField("count", len(e.executors)).Info("killing all")
	return e.each(context.Background(), func(ctx context.Context, exec *executor) error {
		logrus.WithField("droplet.id", exec.droplet.ID).Info("destroying executor")
		_, err := exec.cloud.Droplets.Delete(exec.droplet.ID)
		return err
	})
}

func (e *executors) executeCommand(
	parent context.Context,
	url string,
	script string,
	scriptID string,
	runtime int32,
	maxWorkers int32,
	growthFactor float64,
	timeBetweenGrowth float64,
	startingRPS int32,
	maxRPS int32,
) error {
	return e.each(parent, func(ctx context.Context, exec *executor) error {
		ll := logrus.WithFields(logrus.Fields{
			"droplet.id": exec.droplet.ID,
			"port":       exec.port,
		})
		ll.Info("waiting til executor is alive")
		if err := exec.waitTilAlive(ctx); err != nil {
			return err
		}
		ll.Info("executor is alive")

		if exec.cmdClient == nil {
			cmdCLient, err := exec.client.ExecuteCommand(ctx)
			if err != nil {
				return err
			}
			exec.cmdClient = cmdCLient
		}

		in := &pb.CommandMessage{
			Command: "Run",
			ScriptParams: &pb.ScriptParams{
				Url:                       url,
				Script:                    script,
				ScriptId:                  scriptID,
				RunTime:                   runtime,
				MaxWorkers:                maxWorkers / int32(len(e.executors)),
				GrowthFactor:              growthFactor,
				TimeBetweenGrowth:         timeBetweenGrowth,
				StartingRequestsPerSecond: startingRPS / int32(len(e.executors)),
				MaxRequestsPerSecond:      maxRPS / int32(len(e.executors)),
			},
		}
		ll = ll.WithFields(logrus.Fields{
			"worker.count": in.GetScriptParams().MaxWorkers,
			"max.rps":      in.GetScriptParams().MaxRequestsPerSecond,
			"start.rps":    in.GetScriptParams().StartingRequestsPerSecond,
		})

		ll.Info("sending commands to executor")
		return exec.cmdClient.Send(in)
	})
}

func (e *executors) haltCommand(parent context.Context) error {
	return e.each(parent, func(ctx context.Context, exec *executor) error {
		ll := logrus.WithFields(logrus.Fields{
			"droplet.id": exec.droplet.ID,
			"port":       exec.port,
		})
		if exec.cmdClient == nil {
			return fmt.Errorf("nothing to halt")
		}

		in := &pb.CommandMessage{Command: "Halt"}
		ll.Info("sending commands to executor")
		return exec.cmdClient.Send(in)
	})
}

func (e *executors) waitCompletion(parent context.Context) error {
	return e.each(parent, func(ctx context.Context, exec *executor) error {
		ll := logrus.WithFields(logrus.Fields{
			"droplet.id": exec.droplet.ID,
			"port":       exec.port,
		})
		if exec.cmdClient == nil {
			return fmt.Errorf("no execution running")
		}
		res, err := exec.cmdClient.Recv()
		if err != nil {
			return err
		}
		ll.WithField("status", res.Status).Info("execution completed")
		return nil
	})
}

func (e *executors) each(parent context.Context, fn func(ctx context.Context, exec *executor) error) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	var wg sync.WaitGroup
	errc := make(chan error, len(e.executors))
	logrus.WithField("count", len(e.executors)).Debug("launching parallel requests")
	for _, exec := range e.executors {
		wg.Add(1)
		go func(exec *executor) {
			defer wg.Done()
			logrus.WithField("droplet.id", exec.droplet.ID).Debug("request to executor")
			if err := fn(ctx, exec); err != nil {
				errc <- err
			}
		}(exec)
	}
	go func() {
		logrus.WithField("count", len(e.executors)).Debug("waiting for parallel requests")
		wg.Wait()
		close(errc)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errc:
		return err
	}
}
