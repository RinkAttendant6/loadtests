package executorGRPC

import (
	"git.loadtests.me/loadtests/loadtests/executor/controller"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

// GRPCExecutorStarter this will read what ip to ping from a file
type GRPCExecutorStarter struct {
	persister controller.Persister
}

// NewGRPCExecutorStarter this creates a new GRPCExecutorStarter and sets the directory to look in
func NewGRPCExecutorStarter(persister controller.Persister, port string) (*sync.WaitGroup, *grpc.Server, error) {
	lis, err := net.Listen("tcp", port)
	var wg sync.WaitGroup
	wg.Add(1)
	if err != nil {
		return &wg, nil, err
	}
	executorStarter := &GRPCExecutorStarter{persister}
	s := grpc.NewServer()
	RegisterCommanderServer(s, executorStarter)
	go func() {
		s.Serve(lis)
		wg.Done()
	}()
	return &wg, s, nil
}

// ExecuteCommand is the server interface for listening for a command
func (s *GRPCExecutorStarter) ExecuteCommand(ctx context.Context, in *CommandMessage) (*StatusMessage, error) {
	log.Printf("Go connection: %v", in)
	executorController := controller.Controller{IP: in.IP}
	for i := int32(0); i < in.NumTimes; i++ {
		err := controller.Execute(executorController, s.persister, in.ScriptName)
		if err != nil {
			log.Printf("Error executing: %v", err)
			return &StatusMessage{"Error", err.Error()}, nil
		}
	}
	return &StatusMessage{"OK", ""}, nil
}
