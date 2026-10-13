package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var aiEventPublisherOnce sync.Once
var aiEventPublisherInstance adapter.AIEventPublisher

func GetAIEventPublisher() adapter.AIEventPublisher {
	if aiEventPublisherInstance == nil {
		aiEventPublisherOnce.Do(func() {
			aiEventPublisherInstance = publisher.NewAIEvent(external.GetEventEmitter())
		})
	}
	return aiEventPublisherInstance
}
