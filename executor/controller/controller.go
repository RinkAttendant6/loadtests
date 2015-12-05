package controller

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	client "github.com/influxdb/influxdb/client/v2"
	"github.com/lgpeterson/loadtests/executor/engine"
	"github.com/lgpeterson/loadtests/executor/pb"
)

// Controller this will read what IP to ping from a file
type Controller struct {
	Command *executorGRPC.ScriptParams
	Config  string
	Server  executorGRPC.Commander_ExecuteCommandServer
	Clock   clock.Clock
}

// Persister is an interface to save whatever data is grabbed from the executor
type Persister interface {
	Persist(bps client.BatchPoints) error
	SetupPersister(influxIP string, user string, pass string, database string, useSsl bool) error
}

// RunInstructions will get the IP from the file it found and send it to the pinger
func (f *Controller) RunInstructions(persister Persister, dropletId int, halt chan struct{}) error {
	script := strings.NewReader(f.Command.Script)
	_, err := engine.Lua(script)
	if err != nil {
		return err
	}
	if f.Config != "" {
		cfg := make(map[string]interface{})
		if err = json.Unmarshal([]byte(f.Config), &cfg); err != nil {
			return err
		}
		if err = engine.VerifyConfig(cfg); err != nil {
			return err
		}
	}
	metrics, err := f.runScript(dropletId, halt)
	if err != nil {
		return err
	}
	numTries := 0
	for {
		err := persister.Persist(metrics)
		if err == nil {
			return nil
		}
		log.Println("Failed presist attempt. number: %d err: %v", numTries, err)
		if numTries > 10 {
			return err
		}
	}
}

func (f *Controller) runScript(dropletId int, halt chan struct{}) (client.BatchPoints, error) {
	jobChannel := make(chan struct{}, f.Command.MaxRequestsPerSecond)
	done := make(chan struct{})
	var completeChannels []chan struct{}
	var metricsList []*MetricsGatherer
	var wg sync.WaitGroup

	// Create all the workers that will listen for jobs
	for i := int32(0); i < f.Command.MaxWorkers; i++ {
		workerDone := make(chan struct{})
		metrics, err := NewMetricsGatherer(f.Command.ScriptId, dropletId, i)
		if err != nil {
			return nil, err
		}
		w := &worker{
			WorkerId:   i,
			Command:    f.Command,
			Config:     f.Config,
			Metrics:    metrics,
			Wait:       &wg,
			JobChannel: jobChannel,
			Done:       workerDone,
		}
		completeChannels = append(completeChannels, workerDone)
		metricsList = append(metricsList, metrics)
		go w.execute()
	}

	requestsPerSecond := int(f.Command.StartingRequestsPerSecond)

	// I want to send jobs every 100 miliseconds
	tickTimer := time.Millisecond * 100
	// Find how many jobs to send every tick
	iterations := getNumberOfIterations(tickTimer, requestsPerSecond)

	ticker := f.Clock.Ticker(tickTimer)
	defer ticker.Stop()

	growthTicker := f.Clock.Ticker(time.Second * time.Duration(f.Command.TimeBetweenGrowth))
	defer growthTicker.Stop()
	growthActive := true

	go func() {
		f.Clock.Sleep(time.Second * time.Duration(f.Command.RunTime))
		close(done)
	}()

	for {
	select_again:
		select {
		case <-halt:
			close(jobChannel)
			return stopWorkers(completeChannels, metricsList, &wg)
		case <-done:
			close(jobChannel)
			return stopWorkers(completeChannels, metricsList, &wg)

		case <-ticker.C:
			for i := 1; i < iterations; i++ {
				select {
				case jobChannel <- struct{}{}:
				case <-done:
					break select_again
				}
			}
		case <-growthTicker.C:
			if growthActive {
				requestsPerSecond = int(float64(requestsPerSecond) * f.Command.GrowthFactor)
				if requestsPerSecond > int(f.Command.MaxRequestsPerSecond) {
					// I've now hit the max request per second, so I can't grow anymore
					requestsPerSecond = int(f.Command.MaxRequestsPerSecond)
					growthActive = false
				}
				// The number of jobs per tick will now have increased
				iterations = getNumberOfIterations(tickTimer, requestsPerSecond)
			}
		}
	}

}

func stopWorkers(completeChannels []chan struct{}, metricsList []*MetricsGatherer, wg *sync.WaitGroup) (client.BatchPoints, error) {
	log.Println("Ending load test")
	for _, workerDoneChannel := range completeChannels {
		close(workerDoneChannel)
	}
	wg.Wait()
	conf := client.BatchPointsConfig{}
	bps, err := client.NewBatchPoints(conf)
	if err != nil {
		return nil, err
	}
	for _, metric := range metricsList {
		for _, bp := range metric.BatchPoints.Points() {
			bps.AddPoint(bp)
		}
	}
	return bps, nil
}

func getNumberOfIterations(tickTimer time.Duration, requestsPerSecond int) int {
	return int(float64(requestsPerSecond) * tickTimer.Seconds())
}
