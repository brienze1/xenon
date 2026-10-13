package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var consoleEntriesPublisherOnce sync.Once
var consoleEntriesPublisherInstance adapter.ConsoleEntriesPublisher

func GetConsoleEntriesPublisher() adapter.ConsoleEntriesPublisher {
	if consoleEntriesPublisherInstance == nil {
		consoleEntriesPublisherOnce.Do(func() {
			consoleEntriesPublisherInstance = publisher.NewConsoleEntries(external.GetEventEmitter())
		})
	}
	return consoleEntriesPublisherInstance
}
