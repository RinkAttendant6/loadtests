package persister

import (
	"github.com/influxdb/influxdb/client"
	"os"
)

//import "encoding/json"

// InfluxPersister is a persister that will save the output to a file
type InfluxPersister struct {
	influxIP  string
	tableName string
}

// NewInfluxPersister creates a new influx perisistor with the influx IP
func NewInfluxPersister(influxIP string, tableName string) *InfluxPersister {
	return &InfluxPersister{influxIP, tableName}
}

// Persist saves the data to a file with public permissions
func (f *InfluxPersister) Persist(scriptName string, site string, result []byte) error {
	stringResult := string(result)
	c, err := client.NewClient(&client.ClientConfig{
		Database: "site_development",
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PWD"),
		Host:     f.influxIP,
	})
	if err != nil {
		return err
	}
	series := &client.Series{
		Name:    f.tableName,
		Columns: []string{"scriptName", "site", "result"},
		Points: [][]interface{}{
			{scriptName, site, stringResult},
		},
	}
	if err := c.WriteSeries([]*client.Series{series}); err != nil {
		return err
	}
	return nil
}
