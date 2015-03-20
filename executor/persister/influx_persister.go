package persister

import (
	"log"
)

// InfluxPersister is a persister that will save the output to a file
type InfluxPersister struct {
	influxIP   string
	serverName string
}

// NewInfluxPersister creates a new influx perisistor with the influx IP
func NewInfluxPersister(influxIP string) *InfluxPersister {
	return &InfluxPersister{influxIP, ""}
}

// Persist saves the data to a file with public permissions
func (f *InfluxPersister) Persist(data string) error {
	log.Printf("%s:%s", f.serverName, data)
	// TODO figure out how to send to influx
	return nil
}

// SetScriptName sets what name the output file has
func (f *InfluxPersister) SetScriptName(name string) error {
	f.serverName = name
	return nil
}
