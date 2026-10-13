package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var exportProgressPublisherOnce sync.Once
var exportProgressPublisherInstance adapter.ExportProgressPublisher

func GetExportProgressPublisher() adapter.ExportProgressPublisher {
	if exportProgressPublisherInstance == nil {
		exportProgressPublisherOnce.Do(func() {
			exportProgressPublisherInstance = publisher.NewExportProgress(external.GetEventEmitter())
		})
	}
	return exportProgressPublisherInstance
}
