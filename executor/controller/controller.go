package controller

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
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
	Persist(metrics *MetricsGatherer) error
	SetupPersister(influxIP string, user string, pass string, database string, useSsl bool) error
}

// RunInstructions will get the IP from the file it found and send it to the pinger
func (f *Controller) RunInstructions(persister Persister, halt chan struct{}) error {
	script := strings.NewReader(f.Command.Script)
	_, err := engine.Lua(script)
	if err != nil {
		return err
	}
	if f.Config != "" {
		var j interface{}
		if err = json.Unmarshal([]byte(f.Config), &j); err != nil {
			return err
		}
		if err = engine.VerifyConfig(j); err != nil {
			return err
		}
	}
	f.runScript(persister, halt)
	return nil
}

func (f *Controller) runScript(persister Persister, halt chan struct{}) {
	jobChannel := make(chan struct{}, f.Command.MaxRequestsPerSecond)
	done := make(chan struct{})
	var completeChannels []chan struct{}
	var wg sync.WaitGroup

	// Create all the workers that will listen for jobs
	for i := int32(0); i < f.Command.MaxWorkers; i++ {
		workerDone := make(chan struct{})
		w := &worker{
			WorkerId:   i,
			Persister:  persister,
			Command:    f.Command,
			Config:     f.Config,
			Wait:       &wg,
			JobChannel: jobChannel,
			Done:       workerDone,
		}
		completeChannels = append(completeChannels, workerDone)
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
			stopWorkers(completeChannels, &wg)
			return
		case <-done:
			close(jobChannel)
			stopWorkers(completeChannels, &wg)
			return
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

func stopWorkers(completeChannels []chan struct{}, wg *sync.WaitGroup) {
	log.Println("Ending load test")
	for _, workerDoneChannel := range completeChannels {
		close(workerDoneChannel)
	}
	wg.Wait()
}

func getNumberOfIterations(tickTimer time.Duration, requestsPerSecond int) int {
	return int(float64(requestsPerSecond) * tickTimer.Seconds())
}
