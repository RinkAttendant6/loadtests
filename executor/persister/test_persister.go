package persister

import (
	"fmt"

	"github.com/lgpeterson/loadtests/executor/controller"
)

// TestPersister is a persister that will save the output to a file
type TestPersister struct {
	Content []string
}

// Persist TestPersister the data to a file with public permissions
func (f *TestPersister) Persist(scriptId string, metrics *controller.MetricsGatherer) error {
	for _, point := range metrics.Points {
		data := fmt.Sprintf("%s: %s %d", scriptId, point.Fields["url"], point.Fields["code"])
		f.Content = append(f.Content, data)
	}
	if len(metrics.Points) == 0 {
		data := fmt.Sprintf("%s: ", scriptId)
		f.Content = append(f.Content, data)
	}

	return nil
}
func (f *TestPersister) SetupPersister(influxIP string, user string, pass string, database string, useSsl bool) error {
	return nil
}
