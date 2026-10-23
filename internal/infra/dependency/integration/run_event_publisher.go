package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var runEventPublisherOnce sync.Once
var runEventPublisherInstance adapter.RunEventPublisher

func GetRunEventPublisher() adapter.RunEventPublisher {
	if runEventPublisherInstance == nil {
		runEventPublisherOnce.Do(func() {
			runEventPublisherInstance = publisher.NewRunEvent(external.GetEventEmitter())
		})
	}
	return runEventPublisherInstance
}
