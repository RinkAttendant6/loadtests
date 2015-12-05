package main

import (
	"flag"
	"log"
	"os"
	"time"

	client "github.com/influxdb/influxdb/client/v2"
	"github.com/lgpeterson/loadtests/executor/controller"
	"github.com/lgpeterson/loadtests/executor/persister"
)

func main() {
	log.SetFlags(0)

	addr := flag.String("addr", "localhost:50045", "the IP and port to coneect to")
	flag.Parse()

	testIflux(*addr)
}

func testIflux(ip string) {
	metrics, err := controller.NewMetricsGatherer("12345", 1, 2)
	if err != nil {
		log.Fatalf("Error creating influx persistor: %v", err)
	}
	metrics.IncrHTTPGet("http://localhost/foo", 200, time.Millisecond/10)

	pass := os.Getenv("INFLUX_PWD")
	user := os.Getenv("INFLUX_USER")

	persister := &persister.InfluxPersister{}
	err = persister.SetupPersister(ip, user, pass, "ltm_metrics", true)
	if err != nil {
		log.Fatalf("Error creating influx persistor: %v", err)
	}
	bps := make([]client.BatchPoints, 1)
	bps[0] = metrics.BatchPoints
	err = persister.Persist(bps)
	if err != nil {
		log.Fatalf("Error with influx persistor: %v", err)
	}
	count, err := persister.CountOccurrences("test_run", "GetRequestTable")
	if err != nil {
		log.Fatalf("Error with influx persistor getting count: %v", err)
	}
	log.Printf("Count is: %d", count)

}
