package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher"
)

var workspaceEventPublisherOnce sync.Once
var workspaceEventPublisherInstance adapter.WorkspaceEventPublisher

func GetWorkspaceEventPublisher() adapter.WorkspaceEventPublisher {
	if workspaceEventPublisherInstance == nil {
		workspaceEventPublisherOnce.Do(func() {
			workspaceEventPublisherInstance = publisher.NewWorkspaceEvent(external.GetEventEmitter())
		})
	}
	return workspaceEventPublisherInstance
}
