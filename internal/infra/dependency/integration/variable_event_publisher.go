package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var variableEventPublisherOnce sync.Once
var variableEventPublisherInstance adapter.VariableEventPublisher

func GetVariableEventPublisher() adapter.VariableEventPublisher {
	if variableEventPublisherInstance == nil {
		variableEventPublisherOnce.Do(func() {
			variableEventPublisherInstance = publisher.NewVariableEvent(external.GetEventEmitter())
		})
	}
	return variableEventPublisherInstance
}
