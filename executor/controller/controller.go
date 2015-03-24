package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Controller this will read what IP to ping from a file
type Controller struct {
	IP string
}

// RunInstructions will get the IP from the file it found and send it to the pinger
func (f Controller) RunInstructions(persister Persister) error {
	// The file should only contain the IP that should be pinged
	u, err := url.Parse(strings.TrimSpace(f.IP))
	if err != nil {
		return fmt.Errorf("invalid IP: %v", err)
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return fmt.Errorf("can't fetch IP %q: %v", f.IP, err)
	}
	_ = resp.Body.Close()

	persister.Persist(string(f.IP), strconv.Itoa(resp.StatusCode))

	return nil
}
