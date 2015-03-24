package persister

import (
	"fmt"
	"net/http"
	"strings"
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
func (f *InfluxPersister) Persist(site string, result string) error {
	site = strings.Replace(site, "\"", "", -1)
	str := fmt.Sprintf("[ { \"name\" : \"queryResult\", \"columns\" "+
		": [\"scriptName\", \"siteName\", \"response\"], \"points\" : "+
		" [ [%q, %q, %q] ] } ]", f.serverName, site, result)
	buf := strings.NewReader(str)

	url := fmt.Sprintf("http://%s:50086/db/site_development/series?u=root&p=root", f.influxIP)

	resp, err := http.Post(url, "text/plain", buf)
	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}

// SetScriptName sets what name the output file has
func (f *InfluxPersister) SetScriptName(name string) error {
	f.serverName = name
	return nil
}
