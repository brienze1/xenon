package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var updateStatusPublisherOnce sync.Once
var updateStatusPublisherInstance adapter.UpdateStatusPublisher

func GetUpdateStatusPublisher() adapter.UpdateStatusPublisher {
	if updateStatusPublisherInstance == nil {
		updateStatusPublisherOnce.Do(func() {
			updateStatusPublisherInstance = publisher.NewUpdateStatus(external.GetEventEmitter())
		})
	}
	return updateStatusPublisherInstance
}
