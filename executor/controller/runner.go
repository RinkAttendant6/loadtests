package controller

import (
	"git.loadtests.me/loadtests/loadtests/executor/persister"
)

// ExecutorStarter is an interface that will be used to execute the scripts
type ExecutorStarter interface {
	WaitForInstructions() (string, error)
	RunInstructions(persister persister.Persister) error
}

// Persister is an interface to save whatever data is grabbed from the executor
type Persister interface {
	Persist(data string) error
	SetScriptName(name string) error
}

// Execute is a generic function that take the two interfaces and runs the impletmentation while checking for errors
func Execute(executor ExecutorStarter, persister Persister) error {
	scriptName, err := executor.WaitForInstructions()
	if err != nil {
		return err
	}
	err = persister.SetScriptName(scriptName)
	if err != nil {
		return err
	}
	err = executor.RunInstructions(persister)
	return err
}
