package main

import (
	"git.loadtests.me/loadtests/loadtests/executor/controller"
	"git.loadtests.me/loadtests/loadtests/executor/persister"
	"log"
)

func main() {
	log.SetFlags(0)
	fp := controller.Persister(&persister.FilePersister{})
	ec := controller.ExecutorStarter(controller.NewFileExecutorStarter("./test_data"))

	err := controller.Execute(ec, fp)
	if err != nil {
		log.Fatalf("couldn't execute: %v", err)
	}
	log.Printf("finished!")
}
