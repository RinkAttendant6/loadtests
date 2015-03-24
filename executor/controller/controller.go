package controller

import (
	"bytes"
	"git.loadtests.me/loadtests/loadtests/executor/engine"
	"golang.org/x/net/context"
	"strings"
)

// Controller this will read what IP to ping from a file
type Controller struct {
	IP      string
	Script  string
	Context context.Context
}

// RunInstructions will get the IP from the file it found and send it to the pinger
func (f Controller) RunInstructions(persister Persister) error {
	script := strings.NewReader(f.Script)
	buf := bytes.NewBuffer(nil)
	prog, err := engine.Lua(script, buf)
	if err != nil {
		return err
	}
	err = prog.Execute(f.Context)
	if err != nil {
		return err
	}
	persister.Persist(string(f.IP), buf.String())
	return nil
}
