package controller

import (
	"git.loadtests.me/loadtests/loadtests/executor/executorGRPC"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

// GRPCExecutorStarter this will read what ip to ping from a file
type GRPCExecutorStarter struct {
	persister Persister
}

// NewGRPCExecutorStarter this creates a new GRPCExecutorStarter and sets the directory to look in
func NewGRPCExecutorStarter(persister Persister, port string) (*sync.WaitGroup, *grpc.Server, error) {
	lis, err := net.Listen("tcp", port)
	var wg sync.WaitGroup
	wg.Add(1)
	if err != nil {
		return &wg, nil, err
	}
	executorStarter := &GRPCExecutorStarter{persister}
	s := grpc.NewServer()
	executorGRPC.RegisterCommanderServer(s, executorStarter)
	go func() {
		s.Serve(lis)
		wg.Done()
	}()
	return &wg, s, nil
}

// ExecuteCommand is the server interface for listening for a command
func (s *GRPCExecutorStarter) ExecuteCommand(ctx context.Context, in *executorGRPC.CommandMessage) (*executorGRPC.StatusMessage, error) {
	executorController := &Controller{Command: in, Context: ctx}
	err := Execute(executorController, s.persister, in.ScriptName)
	if err != nil {
		log.Printf("Error executing: %v", err)
		return &executorGRPC.StatusMessage{"Error", err.Error()}, nil
	}
	return &executorGRPC.StatusMessage{"OK", ""}, nil
}
