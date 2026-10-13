package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var gitEventPublisherOnce sync.Once
var gitEventPublisherInstance adapter.GitEventPublisher

func GetGitEventPublisher() adapter.GitEventPublisher {
	if gitEventPublisherInstance == nil {
		gitEventPublisherOnce.Do(func() {
			gitEventPublisherInstance = publisher.NewGitEvent(external.GetEventEmitter())
		})
	}
	return gitEventPublisherInstance
}
