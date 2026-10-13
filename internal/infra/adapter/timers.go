package adapter

import (
	"time"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type timers struct{}

func NewTimers() adapter.Timers {
	return &timers{}
}

func (t *timers) AfterFunc(delay time.Duration, run func()) func() bool {
	return time.AfterFunc(delay, run).Stop
}
