package persister

import (
	"fmt"

	"github.com/lgpeterson/loadtests/executor/controller"
)

// TestPersister is a persister that will save the output to a file
type TestPersister struct {
	GetRequestContent []string
	LoggingContent    []string
}

// Persist TestPersister the data to a file with public permissions
func (f *TestPersister) Persist(metrics *controller.MetricsGatherer) error {
	for _, point := range metrics.BatchPoints.Points() {
		if point.Name() == "GetRequestTable" {
			//fmt.Printf("%v\n", point.Fields())
			data := fmt.Sprintf("%s: %s %d", point.Fields()["id"], point.Fields()["url"], point.Fields()["code"])
			f.GetRequestContent = append(f.GetRequestContent, data)
		} else {
			data := fmt.Sprintf("%s: %v", point.Fields()["id"], point.Fields())
			f.LoggingContent = append(f.LoggingContent, data)
		}
	}

	return nil
}
func (f *TestPersister) SetupPersister(influxIP string, user string, pass string, database string, useSsl bool) error {
	return nil
}
