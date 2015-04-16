package main

import (
	"fmt"
	"github.com/benbjohnson/clock"
	"github.com/lgpeterson/loadtests/executor/controller"
	exgrpc "github.com/lgpeterson/loadtests/executor/executorGRPC"
	"github.com/lgpeterson/loadtests/executor/persister"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net/http"
	"net/http/httptest"
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
	if err != nil {
		t.Fatalf("Error from grpc: %v", err)
	}
	// Create the time when it should be done
	doneTime := clock.NewMock()
	// It should be done in 10 seconds
	doneTime.Add((10 * time.Second) + time.Second)

	time.AfterFunc(time.Second*5, func() { panic("too long") })

	// Mock time passage, every 10ms
	go func() {
		defer s.Stop()
		for timeMock.Now().Before(doneTime.Now()) {
			timeMock.Add(time.Millisecond * 10)
		}

		// Validate responses
		if r.Status == "OK" {
			verifyResults(server, t, &gp)
		} else {
			t.Fatalf("Recieved error when executing: %s", r.Status)
		}
	}()
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

func startServer(t *testing.T, gp controller.Persister, timeMock clock.Clock) (*sync.WaitGroup, *grpc.Server) {
	// Loop forever, because I will wait for commands from the grpc server
	wg := sync.WaitGroup{}
	s, err := controller.NewGRPCExecutorStarter(gp, ":50053", &wg, timeMock)
	if err != nil {
		t.Errorf("err starting grpc server %v", err)
	}
	return &wg, s
}

func sendMesage(message *exgrpc.CommandMessage) (*exgrpc.StatusMessage, error) {
	option := grpc.WithTimeout(15 * time.Second)
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50053", option)
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	c := exgrpc.NewCommanderClient(conn)

	return c.ExecuteCommand(context.Background(), message)
}

func testInflux(t *testing.T) {
	p, err := persister.NewInfluxPersister("45.55.129.22:50086", "root", "root")
	wg, s := startServer(t, p, clock.New())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte("test"))
		case "POST":
		}
	}))
	defer srv.Close()

	script := fmt.Sprintf(`
step.first_step = function()
	resp = get(%q)
end
`, "srv.URL")

	if err != nil {
		t.Errorf("Error createing persistor: %v", err)
	}

	_, err = sendMesage(&exgrpc.CommandMessage{
		URL:                       srv.URL,
		Script:                    script,
		ScriptName:                "test",
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	})
	if err != nil {
		t.Fatalf("Error from grpc: %v", err)
	}
	time.Sleep(time.Second * 20)
	s.Stop()
	wg.Wait()

}
func testActiveServa(t *testing.T) {
	option := grpc.WithTimeout(15 * time.Second)
	// Set up a connection to the server.
	conn, err := grpc.Dial("45.55.159.125:50051", option)
	defer conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	c := exgrpc.NewCommanderClient(conn)
	message := &exgrpc.CommandMessage{
		URL: "htt[://45.55.176.206",
		Script: `step.first_step = function()
	resp = get(http://45.55.176.206)
end`,
		ScriptName:                "test",
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	}
	_, err = c.ExecuteCommand(context.Background(), message)
	if err != nil {
		t.Fatalf("Error from grpc: %v", err)
	}

}
