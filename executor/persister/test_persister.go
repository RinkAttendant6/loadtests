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
func (f *TestPersister) Persist(scriptName string, metrics *controller.MetricsGatherer) error {
	for _, point := range metrics.Points {
		data := fmt.Sprintf("%s: %s %d", scriptName, point.Fields["url"], point.Fields["code"])
		f.Content = append(f.Content, data)
	}
	if len(metrics.Points) == 0 {
		data := fmt.Sprintf("%s: ", scriptName)
		f.Content = append(f.Content, data)
	}

	return nil
}
