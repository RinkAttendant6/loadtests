package controller

import (
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
	m.BatchPoints.AddPoint(client.NewPoint("GetRequestTable",
		nil,
		map[string]interface{}{
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds()},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrHTTPPost(url string, code int, duration time.Duration) {
	m.BatchPoints.AddPoint(client.NewPoint("PostRequestTable",
		nil,
		map[string]interface{}{
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds()},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrHTTPError(url string) {
	m.BatchPoints.AddPoint(client.NewPoint("ErrorRequestTable",
		nil,
		map[string]interface{}{
			"url": url},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrLogInfo(msg interface{}) {
}
func (m *MetricsGatherer) IncrLogFatal(msg interface{}) {
}
