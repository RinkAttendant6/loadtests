package controller

import (
	"time"

	client "github.com/influxdb/influxdb/client/v2"
)

type MetricsGatherer struct {
	BatchPoints client.BatchPoints
	ScriptId    string
}

func NewMetricsGatherer(scriptId string) (*MetricsGatherer, error) {
	conf := client.BatchPointsConfig{}
	bps, err := client.NewBatchPoints(conf)
	if err != nil {
		return nil, err
	}
	return &MetricsGatherer{BatchPoints: bps, ScriptId: scriptId}, nil
}

func (m *MetricsGatherer) IncrScriptExecution() {
	m.BatchPoints.AddPoint(client.NewPoint("ExecutionExecutionTable",
		nil,
		map[string]interface{}{"id": m.ScriptId},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrStepExecution(step string, dur time.Duration) {
	m.BatchPoints.AddPoint(client.NewPoint("StepExecutionTable",
		nil,
		map[string]interface{}{
			"id":          m.ScriptId,
			"duration_ns": dur.Nanoseconds(),
			"step":        step,
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrStepError(step string) {
	m.BatchPoints.AddPoint(client.NewPoint("StepErrorTable",
		nil,
		map[string]interface{}{
			"id":   m.ScriptId,
			"step": step,
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrHTTPGet(url string, code int, duration time.Duration) {
	m.BatchPoints.AddPoint(client.NewPoint("GetRequestTable",
		nil,
		map[string]interface{}{
			"id":          m.ScriptId,
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds(),
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrHTTPPost(url string, code int, duration time.Duration) {
	m.BatchPoints.AddPoint(client.NewPoint("PostRequestTable",
		nil,
		map[string]interface{}{
			"id":          m.ScriptId,
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds(),
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrHTTPError(url string) {
	m.BatchPoints.AddPoint(client.NewPoint("ErrorRequestTable",
		nil,
		map[string]interface{}{
			"id":  m.ScriptId,
			"url": url,
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrLogInfo(msg interface{}) {
	m.logMsg(msg, "info")
}
func (m *MetricsGatherer) IncrLogFatal(msg interface{}) {
	m.logMsg(msg, "fatal")
}

func (m *MetricsGatherer) AddLuaError(err error) {
	m.BatchPoints.AddPoint(client.NewPoint("LuaErrorTable",
		nil,
		map[string]interface{}{
			"id":    m.ScriptId,
			"error": err.Error(),
		},
		time.Now(),
	))
}
func (m *MetricsGatherer) logMsg(msg interface{}, level string) {
	m.BatchPoints.AddPoint(client.NewPoint("LogTable",
		nil,
		map[string]interface{}{
			"id":    m.ScriptId,
			"msg":   msg,
			"level": level,
		},
		time.Now(),
	))
}
