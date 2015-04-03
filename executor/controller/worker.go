package controller

import (
	"bytes"
	"github.com/lgpeterson/loadtests/executor/engine"
	"github.com/lgpeterson/loadtests/executor/executorGRPC"
	"golang.org/x/net/context"
	"io"
	"log"
	"sync"
)

type worker struct {
	Persist    Persister
	Command    *executorGRPC.CommandMessage
	Context    context.Context
	Script     io.Reader
	Wait       *sync.WaitGroup
	JobChannel <-chan int
	EndChannel <-chan bool
}

func (w *worker) execute() {
	w.Wait.Add(1)
	for {
		select {
		case <-w.EndChannel:
			w.Wait.Done()
			return
		case <-w.JobChannel:
			buf := bytes.NewBuffer(nil)
			prog, err := engine.Lua(w.Script, buf)
			if err != nil {
				log.Printf("Error creating lua script: %v", err)
				w.Wait.Done()
				return
			}
			err = prog.Execute(w.Context)

			if err != nil {
				log.Printf("Error running lua script: %v", err)
				continue
			}
			w.Persist.Persist(w.Command.IP, w.Command.Script)
		}

	}
}
