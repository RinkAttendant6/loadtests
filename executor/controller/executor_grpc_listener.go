package controller

import (
	"fmt"
	"github.com/benbjohnson/clock"
	"github.com/lgpeterson/loadtests/executor/executorGRPC"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"strings"
	"sync"
)

// GRPCExecutorStarter this will read what ip to ping from a file
type GRPCExecutorStarter struct {
	persister Persister
	clock     clock.Clock
}

// NewGRPCExecutorStarter this creates a new GRPCExecutorStarter and sets the directory to look in
func NewGRPCExecutorStarter(persister Persister, port string, wg *sync.WaitGroup, clock clock.Clock) (*grpc.Server, error) {
	listenPort := fmt.Sprintf(":%s", port)
	lis, err := net.Listen("tcp", listenPort)
	wg.Add(1)
	if err != nil {
		return nil, err
	}
	executorStarter := &GRPCExecutorStarter{
		persister: persister,
		clock:     clock,
	}
	s := grpc.NewServer()
	executorGRPC.RegisterCommanderServer(s, executorStarter)
	go func() {
		defer wg.Done()
		err := s.Serve(lis)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Fatalf("Grpc server had an error: %v", err)
		}
	}()
	return s, nil
}

// ExecuteCommand is the server interface for listening for a command
func (s *GRPCExecutorStarter) ExecuteCommand(ctx context.Context, in *executorGRPC.CommandMessage) (*executorGRPC.StatusMessage, error) {
	log.Printf("Received command: %v", in)
	executorController := &Controller{Command: in, Context: ctx, Clock: s.clock}
	err := executorController.RunInstructions(s.persister)
	if err != nil {
		log.Printf("Error executing: %v", err)
		return nil, err
	}
	return &executorGRPC.StatusMessage{"OK"}, nil
}
