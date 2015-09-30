package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/benbjohnson/clock"
	meta "github.com/digitalocean/go-metadata"
	"github.com/lgpeterson/loadtests/executor/controller"
	"github.com/lgpeterson/loadtests/executor/persister"
)

func main() {
	log.SetFlags(0)

	var (
		addr      = flag.String("scheduler_addr", "localhost:50045", "the IP and port to connect to")
		port      = flag.Int("port", 50053, "The port for grpc to listen on")
		dropletId = flag.Int("dropletId", -1, "If you want to override the droplet Id being sent to the scheduler")
	)
	flag.Parse()

	start(*addr, *port, *dropletId)
}
func start(schedulerAddr string, port int, dropletId int) {
	persister := &persister.InfluxPersister{}

	// Loop forever, because I will wait for commands from the grpc server
	if dropletId == -1 {
		id, err := getDropletId()
		if err != nil {
			log.Fatalf("error getting droplet id %v", err)
		}
		dropletId = id
	}
	s, err := controller.NewGRPCExecutorStarter(persister, schedulerAddr, port, dropletId, clock.New())
	if err != nil {
		log.Fatalf("err starting grpc server %v", err)
	}
	lis, err := controller.CreateListenPort(port)
	if err != nil {
		log.Fatalf("err creating listening port %v", err)
	}
	log.Printf("Server started on port %d", port)
	s.Serve(lis)
	lis.Close()
}

func getDropletId() (int, error) {
	opt := meta.WithHTTPClient(&http.Client{Timeout: time.Millisecond * 100})
	c := meta.NewClient(opt)
	return c.DropletID()
}
