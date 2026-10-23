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
	case f.Iterations < 1 && !f.IteratesOverData():
		return invalidRunSetting("iterations", "must be at least 1")
	case f.DelayMs < 0 || f.DelayMs > MaxRunDelayMs:
		return invalidRunSetting("delay_ms", "must be between 0 and 3600000")
	default:
		return nil
	}
}

// IteratesOverData is true when no iteration count was given with a data file: one iteration per data row.
func (f FunctionalSettings) IteratesOverData() bool {
	return f.Iterations == 0 && f.DataFile != ""
}

func (f FunctionalSettings) WithDataIterations(rows int) FunctionalSettings {
	if f.IteratesOverData() {
		f.Iterations = max(1, rows)
	}
	return f
}

func (f FunctionalSettings) Delay() time.Duration {
	return time.Duration(f.DelayMs) * time.Millisecond
}

// DelayBefore is the wait before a row: none before the run's first row, the delay before every later one.
func (f FunctionalSettings) DelayBefore(executed int) time.Duration {
	if executed == 0 {
		return 0
	}
	return f.Delay()
}

func (f FunctionalSettings) StopsAfter(row RunRowResult) bool {
	return f.StopOnError && row.IsError()
}
