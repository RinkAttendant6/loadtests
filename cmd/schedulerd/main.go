package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/digitalocean/go-metadata"
	"github.com/digitalocean/godo"
	"github.com/ianschenck/envflag"
	"github.com/lgpeterson/loadtests/scheduler"
	"github.com/lgpeterson/loadtests/scheduler/pb"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

func main() {

	var (
		executorBinaryFilepath = flag.String("executor.binary.filepath", "", "path where the Go binary for the executor can be found")

		port = flag.Int("port", 0, "port on which to listen for service requests")

		token            = flag.String("do.token", "", "DigitalOcean token to use for API access")
		dropletRegion    = flag.String("droplet.region", "nyc3", "DigitalOcean region to start droplets into")
		dropletSize      = flag.String("droplet.size", "512mb", "DigitalOcean droplet size to start")
		dropletImageSlug = flag.String("droplet.image", "coreos-stable", "DigitalOcean image to boot for droplets")

		maxWaitExecutorOnline = flag.Duration("max.wait.executor.online", 2*time.Minute, "max duration to wait for before giving up on an executor to register itself")
		maxWorkerPerExecutor  = flag.Int("max.worker.per.executor", 1000, "max number of threads scheduled on a single executor")
		maxExecPSPerExecutor  = flag.Int("max.rps.per.executor", 100, "max number of requests per second requests of a single executor")

		influxAddr     = flag.String("influx.addr", "", "address where the influx DB can be found")
		influxUsername = flag.String("influx.username", "", "username to use when connecting to influx DB")
		influxPassword = flag.String("influx.password", "", "password to authenticate with influx DB")
		influxDBName   = flag.String("influx.db.name", "", "name of the influx DB to which metrics are sent")
		influxSSL      = flag.Bool("influx.use.ssl", false, "whether to use SSL when talking to influx DB")
	)
	envflag.StringVar(executorBinaryFilepath, "EXECUTOR_BINARY_FILEPATH", "", "")
	envflag.IntVar(port, "PORT", 0, "")
	envflag.StringVar(token, "DO_TOKEN", "", "")
	envflag.StringVar(dropletRegion, "DROPLET_REGION", "", "")
	envflag.StringVar(dropletSize, "DROPLET_SIZE", "", "")
	envflag.StringVar(dropletImageSlug, "DROPLET_IMAGE", "", "")
	envflag.DurationVar(maxWaitExecutorOnline, "MAX_WAIT_EXECUTOR_ONLINE", 0, "")
	envflag.IntVar(maxWorkerPerExecutor, "MAX_WORKER_PER_EXECUTOR", 0, "")
	envflag.IntVar(maxExecPSPerExecutor, "MAX_RPS_PER_EXECUTOR", 0, "")
	envflag.StringVar(influxAddr, "INFLUX_ADDR", "", "")
	envflag.StringVar(influxUsername, "INFLUX_USERNAME", "", "")
	envflag.StringVar(influxPassword, "INFLUX_PASSWORD", "", "")
	envflag.StringVar(influxDBName, "INFLUX_DB_NAME", "", "")
	envflag.BoolVar(influxSSL, "INFLUX_USE_SSL", false, "")

	envflag.Parse()
	flag.Parse()

	md, err := metadata.NewClient().Metadata()
	if err != nil {
		logrus.WithError(err).Fatal("can't reach DO metadata service")
	}
	iface := md.Interfaces["public"][0].IPv4.IPAddress

	addr := startExecutorBinaryFileserver(iface, *executorBinaryFilepath)
	cloud, sshKeys := connectToDO(*token)

	svcl, err := net.Listen("tcp", fmt.Sprintf("%s:%d", iface, *port))
	if err != nil {
		logrus.WithError(err).Fatal("can't provide listener for scheduler service")
	}
	defer svcl.Close()

	cfg := &scheduler.Config{
		PullExecutorBinaryURL: fmt.Sprintf("http://%s", addr.String()),
		AdvertiseListenAddr:   svcl.Addr().String(),
		SSHKeyIDs:             sshKeys,

		DropletRegion:    *dropletRegion,
		DropletSize:      *dropletSize,
		DropletImageSlug: *dropletImageSlug,

		MaxWaitExecutorOnline: *maxWaitExecutorOnline,
		MaxWorkerPerExecutor:  *maxWorkerPerExecutor,
		MaxExecPSPerExecutor:  *maxExecPSPerExecutor,

		InfluxAddr:     *influxAddr,
		InfluxUsername: *influxUsername,
		InfluxPassword: *influxPassword,
		InfluxDBName:   *influxDBName,
		InfluxSSL:      *influxSSL,
	}

	db, err := scheduler.NewDB(cfg, cloud)
	if err != nil {
		logrus.WithError(err).Fatal("can't prepare DB")
	}
	svc := scheduler.NewServer(cfg, db)
	srv := grpc.NewServer()
	pb.RegisterSchedulerServer(srv, svc)

	logrus.WithField("addr", svcl.Addr().String()).Info("scheduler RPC listening")
	if err := srv.Serve(svcl); err != nil {
		logrus.WithError(err).Fatal("can't service requests")
	}
}

func startExecutorBinaryFileserver(iface, filepath string) *net.TCPAddr {
	_, err := os.Stat(filepath)
	if err != nil {
		logrus.WithError(err).WithField("path", filepath).Fatal("can't serve given path as an executor binary file")
	}

	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", iface))
	if err != nil {
		logrus.WithError(err).Fatal("can't provide listener for executor binary server")
	}
	go func() {
		defer l.Close()
		srv := http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath)
		})}
		logrus.WithField("addr", l.Addr().String()).Info("executor binary provider listening")
		if err := srv.Serve(l); err != nil {
			logrus.WithError(err).Panic("can't serve executor binaries")
		}
	}()

	return l.Addr().(*net.TCPAddr)
}

func connectToDO(token string) (*godo.Client, []godo.DropletCreateSSHKey) {
	cloud := godo.NewClient(oauth2.NewClient(oauth2.NoContext,
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
	))
	var sshKeys []godo.DropletCreateSSHKey
	if keys, _, err := cloud.Keys.List(&godo.ListOptions{}); err != nil {
		logrus.WithError(err).Fatal("can't retrieve SSH keys")
	} else if len(keys) == 0 {
		logrus.Fatal("need at least 1 ssh key on your DO account")
	} else {
		for _, ssh := range keys {
			sshKeys = append(sshKeys, godo.DropletCreateSSHKey{ID: ssh.ID})
		}
	}
	return cloud, sshKeys
}
