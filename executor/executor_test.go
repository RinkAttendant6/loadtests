package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/lgpeterson/loadtests/executor/controller"
	exgrpc "github.com/lgpeterson/loadtests/executor/executorGRPC"
	"github.com/lgpeterson/loadtests/executor/persister"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	goodLogScript = `step.first_step = function()
    info("hello world")
end

step.second_step = function()
    fatal("oh you're still there")
end
`
	goodGetScript = `step.first_step = function()
    get(%q)
end
`
	badScript = `testtesttesttest test`

	defaultPort = "50053"
)

func TestValidLoggingCode(t *testing.T) {
	//TODO remove race condition for test cases
	gp := persister.TestPersister{}

	timeMock := clock.NewMock()
	server := "http://localhost"
	wg, s := startServer(t, &gp, timeMock, defaultPort)
	scriptName := fmt.Sprintf("test: %d", rand.Int63())
	r, err := sendMesage(&exgrpc.CommandMessage{
		ScriptName:                scriptName,
		URL:                       server,
		Script:                    goodLogScript,
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	}, defaultPort)
	if err != nil {
		t.Fatalf("Error from grpc: %v", err)
	}
	// Create the time when it should be done
	doneTime := clock.NewMock()
	// It should be done in 10 seconds
	doneTime.Add((10 * time.Second) + time.Second)

	time.AfterFunc(time.Second*5, func() { panic("too long") })

	for timeMock.Now().Before(doneTime.Now()) {
		timeMock.Add(time.Millisecond * 10)
	}

	// Stop the server and wait for the executor stop finish
	s.Stop()
	wg.Wait()

	// Validate responses
	if r.Status == "OK" {
		verifyResults(scriptName, t, &gp)
	} else {
		t.Fatalf("Received error when executing: %s", r.Status)
	}
}

func TestValidGetCode(t *testing.T) {
	//TODO remove race condition for test cases
	gp := persister.TestPersister{}

	timeMock := clock.NewMock()
	wg, s := startServer(t, &gp, timeMock, defaultPort)
	numReq := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte("test"))
			numReq++
		}
	}))
	defer srv.Close()

	script := fmt.Sprintf(goodGetScript, srv.URL)

	scriptName := fmt.Sprintf("test: %d", rand.Int63())
	r, err := sendMesage(&exgrpc.CommandMessage{
		ScriptName:                scriptName,
		URL:                       srv.URL,
		Script:                    script,
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	}, defaultPort)
	if err != nil {
		t.Fatalf("Error from grpc: %v", err)
	}
	// Create the time when it should be done
	doneTime := clock.NewMock()
	// It should be done in 10 seconds
	doneTime.Add((10 * time.Second) + time.Second)

	// Make sure it doesn't deadlock
	time.AfterFunc(time.Second*50, func() { panic("too long") })

	// Mock time passage, every 10ms
	for timeMock.Now().Before(doneTime.Now()) {
		timeMock.Add(time.Millisecond * 10)
	}

	// Stop the server and wait for the executor stop finish
	s.Stop()
	wg.Wait()
	// Validate responses
	if r.Status == "OK" {
		if numReq == 0 {
			t.Fatal("Received no get requests from script")
		}
		// Check that the script stored the correct test ID
		verifyResults(scriptName, t, &gp)
		// Make sure it got good responses
		verifyResults(fmt.Sprintf("%s %d", srv.URL, 200), t, &gp)
	} else {
		t.Fatalf("Received error when executing: %s", r.Status)
	}
}

func TestInvalidCode(t *testing.T) {
	gp := persister.TestPersister{}

	timeMock := clock.NewMock()
	server := "http://localhost"
	wg, s := startServer(t, &gp, timeMock, defaultPort)
	_, err := sendMesage(&exgrpc.CommandMessage{
		ScriptName:                "test",
		URL:                       server,
		Script:                    badScript,
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	}, defaultPort)
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
			t.Fatalf("Invalid content was looking for %s got: %s", server, gp.Content[i])
		}
	}
}

func startServer(t *testing.T, gp controller.Persister, timeMock clock.Clock, port string) (*sync.WaitGroup, *grpc.Server) {
	// Loop forever, because I will wait for commands from the grpc server
	wg := sync.WaitGroup{}
	s, err := controller.NewGRPCExecutorStarter(gp, port, &wg, timeMock)
	if err != nil {
		t.Errorf("err starting grpc server %v", err)
	}
	return &wg, s
}

func sendMesage(message *exgrpc.CommandMessage, port string) (*exgrpc.StatusMessage, error) {

	option := grpc.WithTimeout(15 * time.Second)
	// Set up a connection to the server.
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", port), option)
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	c := exgrpc.NewCommanderClient(conn)

	return c.ExecuteCommand(context.Background(), message)
}
