package controller

import (
	"sync"
	"time"

	client "github.com/influxdb/influxdb/client/v2"
)

type MetricsGatherer struct {
	BatchPoints client.BatchPoints
	ScriptId    string
	DropletId   int
	WorkerId    int32
	TestId      int
	Mutex       *sync.Mutex
}

func NewMetricsGatherer(scriptId string, dropletId int, workerId int32) (*MetricsGatherer, error) {
	conf := client.BatchPointsConfig{}
	bps, err := client.NewBatchPoints(conf)
	if err != nil {
		return nil, err
	}
	return &MetricsGatherer{BatchPoints: bps, ScriptId: scriptId,
		DropletId: dropletId, WorkerId: workerId, TestId: 0, Mutex: &sync.Mutex{}}, nil
}

func (m *MetricsGatherer) ClearBatchPoints() (client.BatchPoints, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	bps := m.BatchPoints

	conf := client.BatchPointsConfig{}
	newBps, err := client.NewBatchPoints(conf)
	if err != nil {
		return nil, err
	}
	m.BatchPoints = newBps

	return bps, nil
}

func (m *MetricsGatherer) IncrScriptExecution() {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.BatchPoints.AddPoint(client.NewPoint("ExecutionExecutionTable",
		nil,
		map[string]interface{}{
			"serverId": m.DropletId,
			"threadId": m.WorkerId,
			"testId":   m.TestId,
			"id":       m.ScriptId},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrStepExecution(step string, dur time.Duration) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.BatchPoints.AddPoint(client.NewPoint("StepExecutionTable",
		nil,
		map[string]interface{}{
			"serverId":    m.DropletId,
			"threadId":    m.WorkerId,
			"testId":      m.TestId,
			"id":          m.ScriptId,
			"duration_ns": dur.Nanoseconds(),
			"step":        step,
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrStepError(step string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.BatchPoints.AddPoint(client.NewPoint("StepErrorTable",
		nil,
		map[string]interface{}{
			"serverId": m.DropletId,
			"threadId": m.WorkerId,
			"testId":   m.TestId,
			"id":       m.ScriptId,
			"step":     step,
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrHTTPGet(url string, code int, duration time.Duration) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.BatchPoints.AddPoint(client.NewPoint("GetRequestTable",
		nil,
		map[string]interface{}{
			"serverId":    m.DropletId,
			"threadId":    m.WorkerId,
			"testId":      m.TestId,
			"id":          m.ScriptId,
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds(),
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrHTTPPost(url string, code int, duration time.Duration) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.BatchPoints.AddPoint(client.NewPoint("PostRequestTable",
		nil,
		map[string]interface{}{
			"serverId":    m.DropletId,
			"threadId":    m.WorkerId,
			"testId":      m.TestId,
			"id":          m.ScriptId,
			"url":         url,
			"code":        code,
			"duration_ns": duration.Nanoseconds(),
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) IncrHTTPError(url string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.BatchPoints.AddPoint(client.NewPoint("ErrorRequestTable",
		nil,
		map[string]interface{}{
			"serverId": m.DropletId,
			"threadId": m.WorkerId,
			"testId":   m.TestId,
			"id":       m.ScriptId,
			"url":      url,
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
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.BatchPoints.AddPoint(client.NewPoint("LuaErrorTable",
		nil,
		map[string]interface{}{
			"serverId": m.DropletId,
			"threadId": m.WorkerId,
			"testId":   m.TestId,
			"id":       m.ScriptId,
			"error":    err.Error(),
		},
		time.Now(),
	))
}

func (m *MetricsGatherer) logMsg(msg interface{}, level string) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.BatchPoints.AddPoint(client.NewPoint("LogTable",
		nil,
		map[string]interface{}{
			"serverId": m.DropletId,
			"threadId": m.WorkerId,
			"testId":   m.TestId,
			"id":       m.ScriptId,
			"msg":      msg,
			"level":    level,
		},
		time.Now(),
	))
}
