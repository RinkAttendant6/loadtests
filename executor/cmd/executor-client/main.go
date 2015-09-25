package main

import (
	"flag"
	"log"
	"time"

	exgrpc "github.com/lgpeterson/loadtests/executor/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	log.SetFlags(0)

	var (
		addr = flag.String("addr", "localhost:50045", "the IP and port to coneect to")
	)
	flag.Parse()

	testServer(*addr)
}
func testServer(ip string) {
	defaultUrl := "45.55.176.206"
	script := `step.first_step = function()
    info("hello world")
    get("http://45.55.176.206")
end
`
	msg, err := sendTestMesage(&exgrpc.CommandMessage{
		Url:                       defaultUrl,
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

func sendTestMesage(message *exgrpc.CommandMessage, ip string) (*exgrpc.StatusMessage, error) {

	option := grpc.WithTimeout(15 * time.Second)
	sec := grpc.WithInsecure()
	// Set up a connection to the server.
	conn, err := grpc.Dial(ip, option, sec)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := exgrpc.NewCommanderClient(conn)

	return c.ExecuteCommand(context.Background(), message)
}
