package main

import (
	"git.loadtests.me/loadtests/loadtests/executor/controller"
	exgrpc "git.loadtests.me/loadtests/loadtests/executor/executorGRPC"
	"git.loadtests.me/loadtests/loadtests/executor/persister"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"strings"
	"sync"
	"testing"
	"time"
)

func getScript() string {
	return `step.first_step = function()
    info("hello world")
end

step.second_step = function()
    fatal("oh you're still there")
end

step.first_step = function()
    info("hello world")
end`
}

func TestValidServer(t *testing.T) {
	//TODO remove race condition for test cases
	gp := persister.TestPersister{}
	testName := "test"
	server := "http://www.google.com"
	wg, s := startServer(t, &gp)
	r, err := sendMesage(&exgrpc.CommandMessage{
		IP:                        server,
		Script:                    getScript(),
		ScriptName:                testName,
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	})
	time.Sleep(time.Second * 15)
	if err == nil && r.Status != "Error" {
		verifyResults(server, t, &gp)
	} else {
		if err != nil {
			t.Errorf("Error when sending command: %v", err)
		} else {
			t.Errorf("Recieved error when executing: %s", r.Error)
		}
	}
	s.Stop()
	wg.Wait()
}

func verifyResults(server string, t *testing.T, gp *persister.TestPersister) {
	if len(gp.Content) < 1 {
		t.Error("No return")
	}
	for i := 0; i < len(gp.Content); i++ {
		if !strings.Contains(gp.Content[i], server) {
			t.Errorf("Invalid content: %s", gp.Content[i])
		}
	}
}

func startServer(t *testing.T, gp *persister.TestPersister) (*sync.WaitGroup, *grpc.Server) {
	// Loop forever, because I will wait for commands from the grpc server
	wg, s, err := controller.NewGRPCExecutorStarter(gp, ":50052")
	if err != nil {
		t.Errorf("err starting grpc server %v", err)
	}
	return wg, s
}

func sendMesage(message *exgrpc.CommandMessage) (*exgrpc.StatusMessage, error) {
	option := grpc.WithTimeout(15 * time.Second)
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50052", option)
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	c := exgrpc.NewCommanderClient(conn)

	return c.ExecuteCommand(context.Background(), message)
}
