package entity

import "time"

const (
	DefaultRunIterations = 1
	MaxRunDelayMs        = 3600000
)

type FunctionalSettings struct {
	DataFile           string
	Iterations         int
	DelayMs            int
	PersistResponses   bool
	DisableLogs        bool
	StopOnError        bool
	KeepVariableValues bool
	SkipStoredCookies  bool
	SaveCookies        bool
}

func NewFunctionalSettings() FunctionalSettings {
	return FunctionalSettings{Iterations: DefaultRunIterations, KeepVariableValues: true}
}

func (f FunctionalSettings) Validate() error {
	switch {
	case f.Iterations < 1:
		return invalidRunSetting("iterations", "must be at least 1")
	case f.DelayMs < 0 || f.DelayMs > MaxRunDelayMs:
		return invalidRunSetting("delay_ms", "must be between 0 and 3600000")
	default:
		return nil
	}
}

func (f FunctionalSettings) Delay() time.Duration {
	return time.Duration(f.DelayMs) * time.Millisecond
}
