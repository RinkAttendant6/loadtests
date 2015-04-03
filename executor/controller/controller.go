package controller

import (
	"git.loadtests.me/loadtests/loadtests/executor/engine"
	"git.loadtests.me/loadtests/loadtests/executor/executorGRPC"
	"golang.org/x/net/context"
	"strings"
	"sync"
	"time"
)

// Controller this will read what IP to ping from a file
type Controller struct {
	Command *executorGRPC.CommandMessage
	Context context.Context
}

// RunInstructions will get the IP from the file it found and send it to the pinger
func (f *Controller) RunInstructions(persister Persister) error {
	script := strings.NewReader(f.Command.Script)
	_, err := engine.Lua(script, nil)
	if err != nil {
		return err
	}
	go f.runScript(persister)
	return nil
}

func (f *Controller) runScript(persister Persister) {
	script := strings.NewReader(f.Command.Script)

	endChannel := make(chan bool)
	jobChannel := make(chan int)
	var wg sync.WaitGroup

	for i := int32(0); i < f.Command.MaxWorkers; i++ {
		w := &worker{
			Persist:    persister,
			Command:    f.Command,
			Context:    f.Context,
			Script:     script,
			Wait:       &wg,
			JobChannel: jobChannel,
			EndChannel: endChannel,
		}
		go w.execute()
	}

	go func() {
		time.Sleep(time.Second * time.Duration(f.Command.RunTime))
		for i := int32(0); i < f.Command.MaxWorkers+1; i++ {
			endChannel <- true
		}
	}()

	tickChan := time.NewTicker(time.Millisecond * 100).C
	growthTicker := time.NewTicker(time.Second * time.Duration(f.Command.TimeBetweenGrowth))

	growthChan := growthTicker.C
	requestsPerSecond := int(f.Command.StartingRequestsPerSecond)
	for {
		select {
		case <-endChannel:
			wg.Wait()
			close(endChannel)
			return
		case <-tickChan:
			iterations := requestsPerSecond / 10
			for i := 1; i < iterations; i++ {
				jobChannel <- 1
			}
		case <-growthChan:
			requestsPerSecond = int(float64(requestsPerSecond) * f.Command.GrowthFactor)
			if requestsPerSecond > int(f.Command.MaxRequestsPerSecond) {
				requestsPerSecond = int(f.Command.MaxRequestsPerSecond)
				growthTicker.Stop()
			}
		}

	}

}
