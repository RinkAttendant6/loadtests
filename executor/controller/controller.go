package controller

import (
	"fmt"
	"golang.org/x/net/context"
	"strings"
)

// Controller this will read what IP to ping from a file
type Controller struct {
	IP      string
	script  string
	context *context.Context
}

// RunInstructions will get the IP from the file it found and send it to the pinger
func (f Controller) RunInstructions(persister Persister) error {
	script := strings.NewReader(f.script)
	buf := bytes.NewBuffer(nil)
	prog := engine.Lua(script, buf)
	prog.Execute(context)
	persister.Persist(fmt.Sprintf("%q: %d", string(f.IP), buf.String()))

	return nil
}
