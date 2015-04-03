package controller

import (
	"bytes"
	"git.loadtests.me/loadtests/loadtests/executor/engine"
	"golang.org/x/net/context"
	"io"
	"strings"
)

// Controller this will read what IP to ping from a file
type Controller struct {
	IP      string
	Script  string
	Context context.Context
}

// RunInstructions will get the IP from the file it found and send it to the pinger
func (f *Controller) RunInstructions(persister Persister) error {
	script := strings.NewReader(f.Script)
	_, err := engine.Lua(script, nil)
	if err != nil {
		return err
	}
	go f.runScript(persister)
	return nil
}

func (f *Controller) runScript(persister Persister) {
	script := strings.NewReader(f.Script)
	f.execute(script, persister)
	// TODO decide what to do with an error
}
func (f *Controller) execute(script io.Reader, persister Persister) error {
	prog, err := engine.Lua(script, nil)
	buf := bytes.NewBuffer(nil)
	if err != nil {
		return err
	}
	err = prog.Execute(f.Context)
	if err != nil {
		return err
	}
	//todo decide if nessisary
	persister.Persist(string(f.IP), buf.String())
	return nil
}
