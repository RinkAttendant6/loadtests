package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/lgpeterson/loadtests/executor/controller"
	exgrpc "github.com/lgpeterson/loadtests/executor/executorGRPC"
	"github.com/lgpeterson/loadtests/executor/persister"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) == 3 && (os.Args[1] == "-t" || os.Args[1] == "-i") {
		// Test a remote server
		ip := os.Args[2]

		// See what to test
		if os.Args[1] == "-t" {
			testServer(ip)
		} else {
			testIflux(ip)
		}
	} else {
		start()
	}
}
func start() {
	port, influxIP, err := extractParams()
	if err != nil {
		log.Fatalf("Error in getting params: %v", err)
	}
	pass := os.Getenv("INFLUX_PWD")
	user := os.Getenv("INFLUX_USER")
	gp, err := persister.NewInfluxPersister(influxIP, user, pass)
	if err != nil {
		log.Fatalf("Error setting up influx client: %v", err)
	}
	wg := sync.WaitGroup{}
	// Loop forever, because I will wait for commands from the grpc server
	_, err = controller.NewGRPCExecutorStarter(gp, port, &wg, clock.New())
	if err != nil {
		log.Fatalf("err starting grpc server %v", err)
	}
	log.Printf("Server started on port %s", port)
	wg.Wait()
}

func testServer(ip string) {
	defaultUrl := "45.55.176.206"
	script := `step.first_step = function()
    info("hello world")
    get("http://45.55.176.206")
end
`
	msg, err := sendTestMesage(&exgrpc.CommandMessage{
		URL:                       defaultUrl,
		Script:                    script,
		ScriptName:                "test",
		RunTime:                   10,
		MaxWorkers:                100,
		GrowthFactor:              1.5,
		TimeBetweenGrowth:         1,
		StartingRequestsPerSecond: 10,
		MaxRequestsPerSecond:      1000,
	}, ip)
	// Validate responses
	if err != nil {
		log.Fatalf("Error in contacting server: %v", err)
	}
	log.Printf("Message was: %v", msg)
}

func testIflux(ip string) {
	metrics := controller.NewMetricsGatherer()
	metrics.IncrHTTPGet("http://localhost/foo", 99, time.Millisecond)
	pass := os.Getenv("INFLUX_PWD")
	user := os.Getenv("INFLUX_USER")
	persister, err := persister.NewInfluxPersister(ip, user, pass)
	if err != nil {
		log.Fatalf("Error creating influx persistor: %v", err)
	}
	err = persister.Persist("test_run", metrics)
	if err != nil {
		log.Fatalf("Error with influx persistor: %v", err)
	}
	count, err := persister.CountOccurrences("test_run", "GetRequestTable")
	if err != nil {
		log.Fatalf("Error with influx persistor getting count: %v", err)
	}
	log.Printf("Count is: %d", count)

}

func sendTestMesage(message *exgrpc.CommandMessage, ip string) (*exgrpc.StatusMessage, error) {

	option := grpc.WithTimeout(15 * time.Second)
	// Set up a connection to the server.
	conn, err := grpc.Dial(ip, option)
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	c := exgrpc.NewCommanderClient(conn)

	return c.ExecuteCommand(context.Background(), message)
}

func extractParams() (string, string, error) {
	if len(os.Args) != 3 {
		return "", "", fmt.Errorf("incorrect params found, expected: <grpc_listen_port influx_database_IP> or -t <remote_ip> or -i <remote_ip>")
	}
	port := os.Args[1]
	influxIP := os.Args[2]
	return port, influxIP, nil
}
