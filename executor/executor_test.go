package main

import (
	"github.com/benbjohnson/clock"
	"github.com/lgpeterson/loadtests/executor/controller"
	exgrpc "github.com/lgpeterson/loadtests/executor/executorGRPC"
	"github.com/lgpeterson/loadtests/executor/persister"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	goodScript = `step.first_step = function()
    info("hello world")
end

step.second_step = function()
    fatal("oh you're still there")
end

step.first_step = function()
    info("hello world")
end`
	badScript = `testtesttesttest test`
)

func TestInflux(t *testing.T) {
	put := `{"lvl":"info","step":"first_step","msg":"hello world"}`
	p := persister.NewInfluxPersister("45.55.129.22:50086", "test")
	p.Persist("t", "m", []byte(put))

}

func TestValidCode(t *testing.T) {
	//TODO remove race condition for test cases
	gp := persister.TestPersister{}

	timeMock := clock.NewMock()
	testName := "test"
	server := "http://localhost"
	wg, s := startServer(t, &gp, timeMock)
	r, err := sendMesage(&exgrpc.CommandMessage{
		URL:                       server,
		Script:                    goodScript,
		ScriptName:                testName,
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	})
	// Create the time when it should be done
	doneTime := clock.NewMock()
	// It should be done in 10 seconds
	doneTime.Add(10 * time.Second)

	// Mock time passage, every 10ms
	for timeMock.Now().Before(doneTime.Now()) {
		timeMock.Add(time.Millisecond * 10)
	}

	// Validate responses
	if err == nil && r.Status == "OK" {
		verifyResults(server, t, &gp)
	} else {
		if err != nil {
			t.Errorf("Error from grpc: %v", err)
		} else {
			t.Errorf("Recieved error when executing: %s", r.Status)
		}
	}

	s.Stop()
	wg.Wait()
}

func TestInvalidCode(t *testing.T) {
	gp := persister.TestPersister{}

	timeMock := clock.NewMock()
	testName := "test"
	server := "http://localhost"
	wg, s := startServer(t, &gp, timeMock)
	_, err := sendMesage(&exgrpc.CommandMessage{
		URL:                       server,
		Script:                    badScript,
		ScriptName:                testName,
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	})
	// Validate responses
	if err == nil {
		t.Errorf("No error from server")
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

func startServer(t *testing.T, gp *persister.TestPersister, timeMock clock.Clock) (*sync.WaitGroup, *grpc.Server) {
	// Loop forever, because I will wait for commands from the grpc server
	wg := sync.WaitGroup{}
	s, err := controller.NewGRPCExecutorStarter(gp, ":50052", &wg, timeMock)
	if err != nil {
		t.Errorf("err starting grpc server %v", err)
	}
	return &wg, s
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
