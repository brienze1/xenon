package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var eventEmitterOnce sync.Once
var eventEmitterInstance adapter.EventEmitter

func GetEventEmitter() adapter.EventEmitter {
	if eventEmitterInstance == nil {
		eventEmitterOnce.Do(func() {
			eventEmitterInstance = infraAdapter.NewEventEmitter()
		})
	}
	return eventEmitterInstance
}

func SetEventEmitter(emitter adapter.EventEmitter) {
	eventEmitterOnce.Do(func() {})
	eventEmitterInstance = emitter
}
