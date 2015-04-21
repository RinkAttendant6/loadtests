package controller

import (
	"time"

	"github.com/lgpeterson/influxdb/client"
)

type MetricsGatherer struct {
	Points []client.Point
}

func NewMetricsGatherer() *MetricsGatherer {
	return new(MetricsGatherer)
}

func (m *MetricsGatherer) IncrScriptExecution()                             {}
func (m *MetricsGatherer) IncrStepExecution(step string, dur time.Duration) {}
func (m *MetricsGatherer) IncrStepError(step string)                        {}

func (m *MetricsGatherer) IncrHTTPGet(url string, code int, duration time.Duration) {
	point := client.Point{
		Measurement: "GetRequestTable",
		Fields: map[string]interface{}{
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds()},
		Time:      time.Now(),
		Precision: "ms",
	}

	m.Points = append(m.Points, point)
}

func (m *MetricsGatherer) IncrHTTPPost(url string, code int, duration time.Duration) {
	point := client.Point{
		Measurement: "PostRequestTable",
		Fields: map[string]interface{}{
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds()},
		Time:      time.Now(),
		Precision: "ms",
	}

	m.Points = append(m.Points, point)
}

func (m *MetricsGatherer) IncrHTTPError(url string) {
	point := client.Point{
		Measurement: "ErrorRequestTable",
		Fields: map[string]interface{}{
			"url": url},
		Time:      time.Now(),
		Precision: "ms",
	}

	m.Points = append(m.Points, point)
}

func (m *MetricsGatherer) IncrLogInfo()  {}
func (m *MetricsGatherer) IncrLogFatal() {}
