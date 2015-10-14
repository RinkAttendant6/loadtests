package persister

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/lgpeterson/influxdb/client"
	"github.com/lgpeterson/loadtests/executor/controller"
)

// InfluxPersister is a persister that will save the output to a file
type InfluxPersister struct {
	client   *client.Client
	database string
}

// NewInfluxPersister creates a new influx perisistor with the influx IP
func (f *InfluxPersister) SetupPersister(influxIP string, user string, pass string, database string, useSsl bool) error {
	url, err := client.ParseConnectionString(influxIP, useSsl)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	config := client.NewConfig()
	config.Username = user
	config.Password = pass
	config.URL = url
	config.HttpClient = httpClient

	c, err := client.NewClient(config)

	f.database = database
	f.client = c

	return err
}

func (f *InfluxPersister) CountOccurrences(testID string, tableName string) (int, error) {
	cmd := fmt.Sprintf("select count(id) from %s where id=%q", tableName, testID)
	query := client.Query{
		Command:  cmd,
		Database: f.database,
	}
	result, err := f.client.Query(query)
	if err != nil {
		return 0, err
	} else if result.Error() != nil {
		return 0, result.Error()
	}
	res := result.Results
	if len(res) == 0 {
		return 0, fmt.Errorf("no rows found")
	}
	stringCount := fmt.Sprintf("%v", res[0])
	log.Printf("Returned string was: %s", res[0].Series[0].Values[0][1])
	count, err := strconv.ParseInt(stringCount, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("what I got back was not an int: %v", err)
	}
	return int(count), nil
}

func (f *InfluxPersister) DropData(tableName string) error {
	// drop series test
	cmd := fmt.Sprintf("drop series %s", tableName)
	query := client.Query{
		Command:  cmd,
		Database: f.database,
	}
	_, err := f.client.Query(query)
	return err
}

// Persist saves the data to a file with public permissions
func (f *InfluxPersister) Persist(scriptId string, metrics *controller.MetricsGatherer) error {
	addId(metrics, scriptId)
	bps := client.BatchPoints{
		Points:          metrics.Points,
		Database:        f.database,
		RetentionPolicy: "default",
	}
	_, err := f.client.Write(bps)
	return err
}

func addId(metrics *controller.MetricsGatherer, scriptId string) {
	for _, point := range metrics.Points {
		cols := point.Fields
		cols["id"] = scriptId
		point.Fields = cols
	}
}
