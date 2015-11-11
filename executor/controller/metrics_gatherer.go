package controller

import (
	"log"
	"time"

	client "github.com/influxdb/influxdb/client/v2"
)

type MetricsGatherer struct {
	BatchPoints client.BatchPoints
}

func NewMetricsGatherer() (*MetricsGatherer, error) {
	conf := client.BatchPointsConfig{}
	bps, err := client.NewBatchPoints(conf)
	if err != nil {
		return nil, err
	}
	return &MetricsGatherer{BatchPoints: bps}, nil
}

func (m *MetricsGatherer) IncrScriptExecution()                             {}
func (m *MetricsGatherer) IncrStepExecution(step string, dur time.Duration) {}
func (m *MetricsGatherer) IncrStepError(step string)                        {}

func (m *MetricsGatherer) IncrHTTPGet(url string, code int, duration time.Duration) {
	point, err := client.NewPoint("GetRequestTable",
		nil,
		map[string]interface{}{
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds()},
		time.Now(),
	)

	if err != nil {
		log.Printf("Error creating point: ", err)
	}

	m.BatchPoints.AddPoint(point)
}

func (m *MetricsGatherer) IncrHTTPPost(url string, code int, duration time.Duration) {
	point, err := client.NewPoint("PostRequestTable",
		nil,
		map[string]interface{}{
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds()},
		time.Now(),
	)

	if err != nil {
		log.Printf("Error creating point: ", err)
	}

	m.BatchPoints.AddPoint(point)
}

func (m *MetricsGatherer) IncrHTTPError(url string) {
	point, err := client.NewPoint("ErrorRequestTable",
		nil,
		map[string]interface{}{
			"url": url},
		time.Now(),
	)

	if err != nil {
		log.Printf("Error creating point: ", err)
	}

	m.BatchPoints.AddPoint(point)
}

func (m *MetricsGatherer) IncrLogInfo(msg interface{}) {
	log.Println(msg)
}
func (m *MetricsGatherer) IncrLogFatal(msg interface{}) {
	log.Println(msg)
}
