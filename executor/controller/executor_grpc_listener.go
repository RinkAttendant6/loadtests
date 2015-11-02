package controller

import (
	"fmt"
	"log"
	"net"
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
func NewGRPCExecutorStarter(persister Persister, schedulerAddr string, port int, dropletId int, clock clock.Clock) (*grpc.Server, error) {
	err := registerDroplet(dropletId, persister, schedulerAddr, port)
	if err != nil {
		return nil, err
	}

	executorStarter := &GRPCExecutorStarter{
		persister: persister,
		clock:     clock,
	}
	s := grpc.NewServer()
	executor.RegisterCommanderServer(s, executorStarter)
	return s, nil
}

func registerDroplet(dropletId int, persister Persister,
	schedulerAddr string, port int) error {

	req := &scheduler.RegisterExecutorReq{
		Port:      int64(port),
		DropletId: int64(dropletId),
	}

	timeout := grpc.WithTimeout(15 * time.Second)
	insecure := grpc.WithInsecure()
	// Set up a connection to the server.
	conn, err := grpc.Dial(schedulerAddr, timeout, insecure)
	if err != nil {
		return err
	}
	defer conn.Close()
	c := scheduler.NewSchedulerClient(conn)

	msg, err := c.RegisterExecutor(context.Background(), req)
	if err != nil {
		return err
	}

	return persister.SetupPersister(msg.InfluxAddr, msg.InfluxUsername, msg.InfluxPassword, msg.InfluxDb, msg.InfluxSsl)
}

// ExecuteCommand is the server interface for listening for a command
func (s *GRPCExecutorStarter) ExecuteCommand(server executor.Commander_ExecuteCommandServer) error {
	var halt = make(chan struct{})
	var halted = false
	var serverErr error

	in, err := server.Recv()
	if err != nil {
		log.Printf("Error from scheduler: %v", err)
	}
	// Don't trust the user to give me what I want

	if in.Command != "Run" {
		// I will only accept the 'Run' command at this stage
		err = server.Send(&executor.StatusMessage{Status: "Invalid"})
		return err
	}
	if err = verifyCommand(in.ScriptParams); err != nil {
		log.Printf("Invalid Command Given: %v", err)
		return err
	}

	log.Printf("Received command: %v", in)
	executorController := &Controller{Command: in.ScriptParams, Server: server, Clock: s.clock}

	go listenForHalt(halt, &halted, &serverErr, server)

	err = executorController.RunInstructions(s.persister, halt)

	if err != nil {
		log.Printf("Error executing: %v", err)
		return err
	} else if serverErr != nil {
		// If the recv wait gave an error I want to return it, if possible
		return serverErr
	} else if halted {
		// I want to tell the server I halted
		log.Println("Halted")
		err = server.Send(&executor.StatusMessage{Status: "Halted"})
		return err
	} else {
		err = server.Send(&executor.StatusMessage{Status: "OK"})
		return err
	}
}

func listenForHalt(halt chan struct{}, halted *bool, serverErr *error, server executor.Commander_ExecuteCommandServer) {
	defer func() {
		// This function will execute if the connection is closed
		// There is no way to recv with polling, so I resort to catching the panic when the connection closes
		defer close(halt)
		if serverErr := recover(); serverErr != nil {
			fmt.Printf("Recovered from panic: %q \n", serverErr)
		}
	}()
	for {
		mes, serverErr := server.Recv()
		if serverErr != nil {
			log.Printf("err from scheduler: %v", serverErr)
			// If there is an error, I assume it means that the server may not be able to
			// communicate with the executor, and halt the execution
			return
		} else if mes != nil {
			if mes.Command == "Halt" {
				// Stop execution and turn the halted flag on so I know to send the 'Halted' message back
				*halted = true
				log.Println("Halting now")
				return
			} else {
				// I will only accept the 'Halt' command at this stage
				server.Send(&executor.StatusMessage{Status: "Invalid"})
			}
		}
	}
}

func CreateListenPort(port int) (net.Listener, error) {
	listenPort := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", listenPort)
	if err != nil {
		return nil, err
	}
	return lis, nil
}

func verifyCommand(in *executor.ScriptParams) error {

	// TODO find out the max goroutines the executor can handle
	if in.MaxWorkers < 1 {
		return fmt.Errorf("Max Workers must be greater than 0. Given MaxWorkers: %d", in.MaxWorkers)
	}
	if in.RunTime <= 1 {
		return fmt.Errorf("You must run for greater than 1 second. Given Runtime: %d", in.RunTime)
	}
	if in.StartingRequestsPerSecond <= 10 {
		return fmt.Errorf("Max Requests per seconds must be greater than 10. Given StartingRequestsPerSecond: %d", in.StartingRequestsPerSecond)
	}
	// TODO find out the max requests a second the executor can do
	if in.MaxRequestsPerSecond < 1 {
		return fmt.Errorf("Max Requests per seconds must be greater than 0. Given MaxRequestsPerSecond: %d", in.MaxRequestsPerSecond)
	}
	if in.MaxRequestsPerSecond < in.StartingRequestsPerSecond {
		return fmt.Errorf("Max Requests per seconds must be greater than Starting Requests Per Second. Given MaxRequestsPerSecond: %d StartingRequestsPerSecond: %d",
			in.MaxRequestsPerSecond, in.StartingRequestsPerSecond)
	}
	if in.GrowthFactor < 1 {
		return fmt.Errorf("Growth Factor must be greater or equal to 1. Given GrowthFactor: %d", in.GrowthFactor)
	}
	if in.TimeBetweenGrowth < 0.1 {
		return fmt.Errorf("Time Between Growth must be greater or equal to 0.1. Given TimeBetweenGrowth: %d", in.TimeBetweenGrowth)
	}
	return nil
}
