package persister

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	client "github.com/influxdb/influxdb/client/v2"
	"github.com/lgpeterson/loadtests/executor/controller"
)

// InfluxPersister is a persister that will save the output to a file
type InfluxPersister struct {
	client   client.Client
	database string
}

// NewInfluxPersister creates a new influx perisistor with the influx IP
func (f *InfluxPersister) SetupPersister(influxIP string, user string, pass string, database string, useSsl bool) error {
	influxUrl := parseUrl(influxIP, useSsl)
	url, err := url.Parse(influxUrl)
	if err != nil {
		return err
	}

	config := client.Config{
		Username:           user,
		Password:           pass,
		URL:                url,
		InsecureSkipVerify: useSsl,
	}

	c := client.NewClient(config)

	f.database = database
	f.client = c

	return err
}

func (f *InfluxPersister) CountOccurrences(testID string, tableName string) (int, error) {
	cmd := fmt.Sprintf("select count(id) from %s where id='%s'", tableName, testID)
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
	stringCount := fmt.Sprintf("%v", res[0].Series[0].Values[0][1])
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
func (f *InfluxPersister) Persist(metrics *controller.MetricsGatherer) error {
	bps := metrics.BatchPoints
	bps.SetDatabase(f.database)
	err := f.client.Write(bps)
	return err
}

func parseUrl(url string, useSsl bool) string {
	var prefix string

	if useSsl {
		prefix = "https://"
	} else {
		prefix = "http://"
	}

	if !strings.HasPrefix(url, prefix) {
		urls := []string{prefix, url}
		url = strings.Join(urls, "")
	}
	return url
}
