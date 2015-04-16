package main

import (
	"fmt"
	"github.com/benbjohnson/clock"
	"github.com/lgpeterson/loadtests/executor/controller"
	"github.com/lgpeterson/loadtests/executor/persister"
	"log"
	"os"
	"sync"
)

func main() {
	log.SetFlags(0)
	influxIP, err := extractParams()
	if err != nil {
		log.Fatalf("Error in getting params: %v", err)
	}
	pass := os.Getenv("INFLUX_PWD")
	user := os.Getenv("INFLUX_USER")
	gp, err := persister.NewInfluxPersister(influxIP, user, pass)
	if err != nil {
		log.Fatalf("Error setting up influx client: %v", err)
	}
	wg := sync.WaitGroup{}
	// Loop forever, because I will wait for commands from the grpc server
	_, err = controller.NewGRPCExecutorStarter(gp, ":50051", &wg, clock.New())
	if err != nil {
		log.Fatalf("err starting grpc server %v", err)
	}
	log.Println("Server started on port :50051")
	wg.Wait()
}

func extractParams() (string, error) {
	if len(os.Args) != 2 {
		return "", fmt.Errorf("incorrect params found, expected: <influx database IP>")
	}
	influxIP := os.Args[1]
	return influxIP, nil
}
