package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var timersOnce sync.Once
var timersInstance adapter.Timers

func GetTimers() adapter.Timers {
	if timersInstance == nil {
		timersOnce.Do(func() {
			timersInstance = infraAdapter.NewTimers()
		})
	}
	return timersInstance
}

func SetTimers(timers adapter.Timers) {
	timersOnce.Do(func() {})
	timersInstance = timers
}
