package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var importProgressPublisherOnce sync.Once
var importProgressPublisherInstance adapter.ImportProgressPublisher

func GetImportProgressPublisher() adapter.ImportProgressPublisher {
	if importProgressPublisherInstance == nil {
		importProgressPublisherOnce.Do(func() {
			importProgressPublisherInstance = publisher.NewImportProgress(external.GetEventEmitter())
		})
	}
	return importProgressPublisherInstance
}
