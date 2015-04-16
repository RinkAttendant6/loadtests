package persister

import (
	"github.com/influxdb/influxdb/client"
	"log"
	"time"
)

//import "encoding/json"

// InfluxPersister is a persister that will save the output to a file
type InfluxPersister struct {
	client *client.Client
}

// NewInfluxPersister creates a new influx perisistor with the influx IP
func NewInfluxPersister(influxIP string, user string, pass string) (*InfluxPersister, error) {
	c, err := client.NewClient(&client.ClientConfig{
		Database: "site_development",
		Username: user,
		Password: pass,
		Host:     influxIP,
	})
	return &InfluxPersister{c}, err
}

func (f *InfluxPersister) IncrScriptExecution()                             {}
func (f *InfluxPersister) IncrStepExecution(step string, dur time.Duration) {}
func (f *InfluxPersister) IncrStepError(step string)                        {}
func (f *InfluxPersister) IncrHTTPGet(url string, code int, duration time.Duration) {
	series := &client.Series{
		Name:    "GetRequestTable",
		Columns: []string{"url", "code", "duration_ns"},
		Points: [][]interface{}{
			{url, code, duration.Nanoseconds()},
		},
	}
	err := f.client.WriteSeries([]*client.Series{series})
	if err != nil {
		log.Printf("error writing post series to influx: %v", err)
	}
}
func (f *InfluxPersister) IncrHTTPPost(url string, code int, duration time.Duration) {
	series := &client.Series{
		Name:    "PostRequestTable",
		Columns: []string{"url", "code", "duration_ns"},
		Points: [][]interface{}{
			{url, code, duration.Nanoseconds()},
		},
	}
	err := f.client.WriteSeries([]*client.Series{series})
	if err != nil {
		log.Printf("error writing post series to influx: %v", err)
	}
}
func (f *InfluxPersister) IncrHTTPError(url string) {
	series := &client.Series{
		Name:    "ErrorRequestTable",
		Columns: []string{"url"},
		Points: [][]interface{}{
			{url},
		},
	}
	err := f.client.WriteSeries([]*client.Series{series})
	if err != nil {
		log.Printf("error writing post series to influx: %v", err)
	}
}
func (f *InfluxPersister) IncrLogInfo()  {}
func (f *InfluxPersister) IncrLogFatal() {}

// Persist saves the data to a file with public permissions
func (f *InfluxPersister) Persist(scriptName string, site string, result []byte) error {
	stringResult := string(result)

	series := &client.Series{
		Name:    "test",
		Columns: []string{"scriptName", "site", "result"},
		Points: [][]interface{}{
			{scriptName, site, stringResult},
		},
	}
	err := f.client.WriteSeries([]*client.Series{series})
	return err
}
