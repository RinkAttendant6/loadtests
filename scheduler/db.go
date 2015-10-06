package scheduler

import (
	"fmt"
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

curl %q > /opt/executord
chmod +x /opt/executord

/etc/systemd/system/load_executor.service <<EOF
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

	return &DB{cloud: cloud, waitDroplets: make(map[int]chan<- int)}, nil
}

func (db *DB) LaunchExecutors(ctx context.Context, count int) (*executors, error) {
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
			port := make(chan int, 1)
			db.waitDroplets[droplet.ID] = port
			db.lock.Unlock()
			defer func() {
				db.lock.Lock()
				delete(db.waitDroplets, droplet.ID)
				db.lock.Unlock()
			}()

			select {
			case <-port:

			case <-ctx.Done():
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

func (db *DB) RegisterExecutorUp(dropletID int, port int) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	wait, ok := db.waitDroplets[dropletID]
	if !ok {
		_, err := db.cloud.Droplets.Delete(dropletID)
		return fmt.Errorf("unexpected droplet %d registered, delete request sent: %v", dropletID, err)
	}
	wait <- port
	return nil
}

type executor struct {
	cloud   *godo.Client
	droplet *godo.Droplet
	port    int
	client  pb.CommanderClient
}

func (e *executor) waitTilAlive(ctx context.Context) error {

	for len(e.droplet.Networks.V4) == 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		var err error
		e.droplet, _, err = e.cloud.Droplets.Get(e.droplet.ID)
		if err != nil {
			return err
		}
	}

	ip := e.droplet.Networks.V4[0].IPAddress
	if len(e.droplet.Networks.V6) != 0 {
		ip = e.droplet.Networks.V6[0].IPAddress
	}
	url := fmt.Sprintf("%s:%d", ip, e.port)

	for {
		if e.client != nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		cc, err := grpc.Dial(url, grpc.WithBlock(), grpc.WithTimeout(time.Second))
		switch err {
		case grpc.ErrClientConnTimeout:
			continue
		case nil:
			e.client = pb.NewCommanderClient(cc)
		}
	}
}

type executors struct {
	cloud     *godo.Client
	executors []*executor
}

func (e *executors) killall() error {
	return e.each(context.Background(), func(ctx context.Context, exec *executor) error {
		_, err := e.cloud.Droplets.Delete(exec.droplet.ID)
		return err
	})
}

func (e *executors) executeCommand(
	parent context.Context,
	url string,
	script string,
	scriptName string,
	runtime int32,
	maxWorkers int32,
	growthFactor float64,
	timeBetweenGrowth float64,
	startingRPS int32,
	maxRPS int32,
) error {
	return e.each(parent, func(ctx context.Context, exec *executor) error {
		if err := exec.waitTilAlive(ctx); err != nil {
			return err
		}

		resp, err := exec.client.ExecuteCommand(ctx, &pb.CommandMessage{
			Url:                       url,
			Script:                    script,
			ScriptName:                scriptName,
			RunTime:                   runtime,
			MaxWorkers:                maxWorkers / int32(len(e.executors)),
			GrowthFactor:              growthFactor,
			TimeBetweenGrowth:         timeBetweenGrowth,
			StartingRequestsPerSecond: startingRPS / int32(len(e.executors)),
			MaxRequestsPerSecond:      maxRPS / int32(len(e.executors)),
		})
		if err != nil {
			return err
		}
		if resp.Status != "OK" {
			return fmt.Errorf("executor %v is not OK: %v", exec.droplet.ID, resp.Status)
		}
		return nil
	})
}

func (e *executors) each(parent context.Context, fn func(ctx context.Context, exec *executor) error) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	var wg sync.WaitGroup
	errc := make(chan error, len(e.executors))
	for _, exec := range e.executors {
		wg.Add(1)
		go func(exec *executor) {
			defer wg.Done()
			if err := fn(ctx, exec); err != nil {
				errc <- err
			}
		}(exec)
	}
	go func() {
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
