package engine

import (
	"golang.org/x/net/context"
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
