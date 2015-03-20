package main

import (
	exgrpc "git.loadtests.me/loadtests/loadtests/executor/executorGRPC"
	"git.loadtests.me/loadtests/loadtests/executor/persister"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestValidServer(t *testing.T) {
	gp := persister.TestPersister{}
	testName := "test"
	server := "http://www.google.com"
	wg, s := startServer(t, &gp)
	r, err := sendMesage(&exgrpc.CommandMessage{IP: server, NumTimes: 2, ScriptName: testName})
	if err == nil && r.Status != "Error" {
		verifyResults(2, server, "200", testName, t, &gp)
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

func TestInvalidServerPage(t *testing.T) {
	gp := persister.TestPersister{}
	testName := "failurePageTest"
	server := "http://www.google.com/errorPageTest1245"
	wg, s := startServer(t, &gp)
	r, err := sendMesage(&exgrpc.CommandMessage{IP: server, NumTimes: 2, ScriptName: testName})
	if err == nil && r.Status != "Error" {
		verifyResults(2, server, "404", testName, t, &gp)
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

func verifyResults(numSent int, server string, status string, testName string, t *testing.T, gp *persister.TestPersister) {
	if len(gp.Content) != numSent {
		t.Errorf("Results were not added, %d entries found", len(gp.Content))
	}
	for i := 0; i < len(gp.Content); i++ {
		if gp.TestName != testName ||
			!strings.Contains(gp.Content[i], server) ||
			!strings.Contains(gp.Content[i], status) {
			t.Errorf("Invalid content: %s", gp.Content[i])
		}
	}
}

func TestInvalidServer(t *testing.T) {
	gp := persister.TestPersister{}
	testName := "test"
	server := "not_a_url"
	wg, s := startServer(t, &gp)

	r, err := sendMesage(&exgrpc.CommandMessage{IP: server, NumTimes: 2, ScriptName: testName})

	if err != nil {
		t.Errorf("Error when sending command: %v", err)
	} else if r.Status != "Error" {
		t.Error("No error when executing")
	}
	s.Stop()
	wg.Wait()
}

func startServer(t *testing.T, gp *persister.TestPersister) (*sync.WaitGroup, *grpc.Server) {
	// Loop forever, because I will wait for commands from the grpc server
	wg, s, err := exgrpc.NewGRPCExecutorStarter(gp, ":50051")
	if err != nil {
		t.Errorf("err starting grpc server %v", err)
	}
	return wg, s
}
func sendMesage(message *exgrpc.CommandMessage) (*exgrpc.StatusMessage, error) {
	option := grpc.WithTimeout(15 * time.Second)
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50051", option)
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	c := exgrpc.NewCommanderClient(conn)

	return c.ExecuteCommand(context.Background(), message)
}
