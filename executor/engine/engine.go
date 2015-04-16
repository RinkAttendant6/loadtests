package engine

import (
	"golang.org/x/net/context"
	"time"
)

type Program interface {
	Execute(context.Context) error
}

type StepError struct {
	Step string
	Err  error
}

func (s *StepError) Error() string {
	return s.Step + ":" + s.Err.Error()
}

type MetricReporter interface {
	IncrScriptExecution()

	IncrStepExecution(string, time.Duration)
	IncrStepError(string)

	IncrHTTPGet(string, int, time.Duration)
	IncrHTTPPost(string, int, time.Duration)
	IncrHTTPError(string)

	IncrLogInfo()
	IncrLogFatal()
}

type nullMetric struct{}

func (_ nullMetric) IncrScriptExecution()                    {}
func (_ nullMetric) IncrStepExecution(string, time.Duration) {}
func (_ nullMetric) IncrStepError(string)                    {}
func (_ nullMetric) IncrHTTPGet(string, int, time.Duration)  {}
func (_ nullMetric) IncrHTTPPost(string, int, time.Duration) {}
func (_ nullMetric) IncrHTTPError(string)                    {}
func (_ nullMetric) IncrLogInfo()                            {}
func (_ nullMetric) IncrLogFatal()                           {}
