package controller

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	executor "github.com/lgpeterson/loadtests/executor/pb"
	scheduler "github.com/lgpeterson/loadtests/scheduler/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// GRPCExecutorStarter this will read what ip to ping from a file
type GRPCExecutorStarter struct {
	persister Persister
	clock     clock.Clock
}

// NewGRPCExecutorStarter this creates a new GRPCExecutorStarter and sets the directory to look in
func NewGRPCExecutorStarter(persister Persister, schedulerIp string, port int, dropletId int, wg *sync.WaitGroup, clock clock.Clock) (*grpc.Server, error) {
	err := registerDroplet(dropletId, persister, schedulerIp, port)
	if err != nil {
		return nil, err
	}

	listenPort := fmt.Sprintf(":%d", port)
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
	executor.RegisterCommanderServer(s, executorStarter)
	go func() {
		defer wg.Done()
		err := s.Serve(lis)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Fatalf("Grpc server had an error: %v", err)
		}
	}()
	return s, nil
}

func registerDroplet(dropletId int, persister Persister,
	schedulerIp string, port int) error {

	req := &scheduler.RegisterExecutorReq{
		Port:      int64(port),
		DropletId: int64(dropletId),
	}

	timeout := grpc.WithTimeout(15 * time.Second)
	insecure := grpc.WithInsecure()
	// Set up a connection to the server.
	conn, err := grpc.Dial(schedulerIp, timeout, insecure)
	if err != nil {
		return err
	}
	defer conn.Close()
	c := scheduler.NewSchedulerClient(conn)

	msg, err := c.RegisterExecutor(context.Background(), req)
	if err != nil {
		return err
	}

	return persister.SetupPersister(msg.InfluxIpPort, msg.InfluxUsername, msg.InfluxPassword, msg.InfluxDb, msg.InfluxSsl)
}

// ExecuteCommand is the server interface for listening for a command
func (s *GRPCExecutorStarter) ExecuteCommand(ctx context.Context, in *executor.CommandMessage) (*executor.StatusMessage, error) {
	log.Printf("Received command: %v", in)
	executorController := &Controller{Command: in, Context: ctx, Clock: s.clock}
	err := executorController.RunInstructions(s.persister)
	if err != nil {
		log.Printf("Error executing: %v", err)
		return nil, err
	}
	return &executor.StatusMessage{Status: "OK"}, nil
}
