package controller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// FileExecutorStarter this will read what ip to ping from a file
type FileExecutorStarter struct {
	testDir string
	file    string
}

// NewFileExecutorStarter this creates a new FileExecutorStarter and sets the directory to look in
func NewFileExecutorStarter(dir string) *FileExecutorStarter {
	return &FileExecutorStarter{dir, ""}
}

// WaitForInstructions will wait for a file to be created in the testDir, before reading it and running it
func (f *FileExecutorStarter) WaitForInstructions() (string, error) {
	if f.file != "" {
		return "", fmt.Errorf("file: %s already executed", f.file)
	}

	// Locate a file with the ip/url to test
	fileInfo, err := ioutil.ReadDir(f.testDir)
	if err != nil {
		return "", err
	}

	for _, file := range fileInfo {
		// Make sure this file is not an output file
		if !strings.HasSuffix(file.Name(), ".out") {
			f.file = fmt.Sprintf("%s/%s", f.testDir, file.Name())
			return f.file, nil
		}
	}
	return "", fmt.Errorf("no file found in %s", f.testDir)
}

// RunInstructions will get the IP from the file it found and send it to the pinger
func (f *FileExecutorStarter) RunInstructions(persister Persister) error {
	// The file should only contain the ip that should be pinged
	ip, err := ioutil.ReadFile(f.file)
	if err != nil {
		return err
	}
	u, err := url.Parse(strings.TrimSpace(string(ip)))
	if err != nil {
		return fmt.Errorf("invalid ip: %v", err)
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return fmt.Errorf("can't fetch ip %q: %v", string(ip), err)
	}
	_ = resp.Body.Close()

	persister.Persist(fmt.Sprintf("%q: %d", string(ip), resp.StatusCode))

	return nil
}
