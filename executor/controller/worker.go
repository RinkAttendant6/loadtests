package controller

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/lgpeterson/loadtests/executor/engine"
	"github.com/lgpeterson/loadtests/executor/pb"
	"golang.org/x/net/context"
)

type worker struct {
	WorkerId   int32
	Persister  Persister
	Config     string
	Command    *executorGRPC.ScriptParams
	Wait       *sync.WaitGroup
	JobChannel <-chan struct{}
	Done       <-chan struct{}
}

func (w *worker) execute() {
	w.Wait.Add(1)
	defer log.Printf("Worker: %d closed", w.WorkerId)
	defer w.Wait.Done()
	for {
		select {
		case <-w.Done:
			return
		case _, ok := <-w.JobChannel:
			// Make sure that the channels are not closed
			if !ok {
				return
			}
			select {
			case <-w.Done:
				return
			default:
			}
			scriptReader := strings.NewReader(w.Command.Script)
			metrics, err := NewMetricsGatherer(w.Command.ScriptId)
			if err != nil {
				// This should not happen, because there are no parameters to NewMetricsGatherer
				// But I should log it for testing/debugging purposes
				log.Printf("Worker %d, Error creating metrics gatherer: %v", w.WorkerId, err)
				return
			}
			prog, err := engine.Lua(scriptReader, engine.SetMetricReporter(metrics))
			if err != nil {
				// This should not be because the script did not compile, if it
				// did not compile it would be reported to the user before this
				log.Printf("Worker %d, Error creating lua script: %v", w.WorkerId, err)
				return
			}
			// Add a config if it exists
			if w.Config != "" {
				cfg := make(map[string]interface{})
				if err = json.Unmarshal([]byte(w.Config), &cfg); err != nil {
					// This should not occur, it would have been checked in controller
					log.Printf("Worker %d, Error unmarshalling: %v", w.WorkerId, err)
					return
				}
				if err = prog.AddConfig(cfg); err != nil {
					log.Printf("Worker %d, Error addding Config to lua script: %v", w.WorkerId, err)
					return
				}
			}
			err = prog.Execute(context.Background())

			if err != nil {
				// I assume I can keep going if the lua script encoutered an error
				metrics.AddLuaError(err)
			}

			err = w.Persister.Persist(metrics)
			if err != nil {
				// I assume I can keep going if the lua script encoutered an error
				log.Printf("Worker %d, Error saving output of lua script: %v", w.WorkerId, err)
			}
		}

	}
}
