package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/lgpeterson/loadtests/executor/controller"
	exgrpc "github.com/lgpeterson/loadtests/executor/pb"
	"github.com/lgpeterson/loadtests/executor/persister"
	scheduler "github.com/lgpeterson/loadtests/scheduler/pb"
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

	defaultPort   = 50053
	schedulerIP   = "localhost:50048"
	schedulerPort = ":50048"
	dropletId     = 125446
)

func TestValidLoggingCode(t *testing.T) {
	//TODO remove race condition for test cases
	gp := persister.TestPersister{}

	timeMock := clock.NewMock()
	server := "http://localhost"
	sch, wg2 := startScheduler(t)
	s, wg := startServer(t, &gp, timeMock, defaultPort)
	scriptName := fmt.Sprintf("test: %d", rand.Int63())
	r, err := sendMesage(&exgrpc.CommandMessage{
		ScriptName:                scriptName,
		Url:                       server,
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

	time.AfterFunc(time.Second*50, func() { panic("too long") })

	for timeMock.Now().Before(doneTime.Now()) {
		timeMock.Add(time.Millisecond * 10)
	}

	// Stop the server and wait for the executor stop finish
	sch.Stop()
	s.Stop()
	wg.Wait()
	wg2.Wait()

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
	sch, wg2 := startScheduler(t)
	s, wg := startServer(t, &gp, timeMock, defaultPort)
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
		Url:                       srv.URL,
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
	sch.Stop()
	s.Stop()
	wg.Wait()
	wg2.Wait()
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
	sch, wg2 := startScheduler(t)
	s, wg := startServer(t, &gp, timeMock, defaultPort)
	_, err := sendMesage(&exgrpc.CommandMessage{
		ScriptName:                "test",
		Url:                       server,
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

	sch.Stop()
	s.Stop()
	wg.Wait()
	wg2.Wait()
}

func verifyResults(server string, t *testing.T, gp *persister.TestPersister) {
	if len(gp.Content) < 1 {
		// attempt to wait for it, it might be slow
		time.Sleep(time.Second * 5)
		// Now I check again, if it fails then the code is taking to long
		if len(gp.Content) < 1 {
			t.Error("No return")
		}
	}
	for i := 0; i < len(gp.Content); i++ {
		if !strings.Contains(gp.Content[i], server) {
			t.Fatalf("Invalid content was looking for %s got: %s", server, gp.Content[i])
		}
	}
}

func startServer(t *testing.T, gp controller.Persister, timeMock clock.Clock, port int) (*grpc.Server, *sync.WaitGroup) {
	// Loop forever, because I will wait for commands from the grpc server
	wg := sync.WaitGroup{}
	s, err := controller.NewGRPCExecutorStarter(gp, schedulerIP, port, dropletId, timeMock)
	if err != nil {
		t.Errorf("err starting grpc server %v", err)
	}

	lis, err := controller.CreateListenPort(port)
	if err != nil {
		t.Fatalf("err creating listening port %v", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Serve(lis)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Fatalf("Grpc server had an error: %v", err)
		}
	}()
	return s, &wg
}

func startScheduler(t *testing.T) (*grpc.Server, *sync.WaitGroup) {
	wg := sync.WaitGroup{}
	s := grpc.NewServer()
	sched := &mockScheduler{}
	lis, err := net.Listen("tcp", schedulerPort)
	if err != nil {
		t.Fatalf("Grpc server had an error: %v", err)
	}
	scheduler.RegisterSchedulerServer(s, sched)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Serve(lis)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			t.Fatalf("Grpc server had an error: %v", err)
		}
	}()
	return s, &wg
}

func sendMesage(message *exgrpc.CommandMessage, port int) (*exgrpc.StatusMessage, error) {
	timeout := grpc.WithTimeout(15 * time.Second)
	insecure := grpc.WithInsecure()
	// Set up a connection to the server.
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), timeout, insecure)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := exgrpc.NewCommanderClient(conn)

	return c.ExecuteCommand(context.Background(), message)
}

type mockScheduler struct{}

func (f *mockScheduler) RegisterExecutor(context.Context, *scheduler.RegisterExecutorReq) (*scheduler.RegisterExecutorResp, error) {
	return &scheduler.RegisterExecutorResp{
		InfluxAddr:     "localhost:12345",
		InfluxUsername: "test",
		InfluxPassword: "test",
		InfluxDb:       "test",
		InfluxSsl:      false,
	}, nil
}

func (f *mockScheduler) LoadTest(in *scheduler.LoadTestReq, s scheduler.Scheduler_LoadTestServer) error {
	return nil
}
