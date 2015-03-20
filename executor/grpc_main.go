package main

import (
	"fmt"
	"git.loadtests.me/loadtests/loadtests/executor/controller"
	grpc "git.loadtests.me/loadtests/loadtests/executor/executorGRPC"
	"git.loadtests.me/loadtests/loadtests/executor/persister"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	influxIP, err := extractParams()
	if err != nil {
		log.Fatalf("Error in getting params: %v", err)
	}
	gp := controller.Persister(persister.NewInfluxPersister(influxIP))
	// Loop forever, because I will wait for commands from the grpc server
	wg, _, err := grpc.NewGRPCExecutorStarter(gp, ":50051")
	if err != nil {
		log.Fatalf("err starting grpc server %v", err)
	}
	wg.Wait()
}

func extractParams() (string, error) {
	if len(os.Args) != 2 {
		return "", fmt.Errorf("incorrect params found, expected: <influx database IP>")
	}
	influxIP := os.Args[1]
	return influxIP, nil
}
