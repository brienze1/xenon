package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var workspaceWatcherConsumerOnce sync.Once
var workspaceWatcherConsumerInstance adapter.WorkspaceWatcherConsumer

func GetWorkspaceWatcherConsumer() adapter.WorkspaceWatcherConsumer {
	if workspaceWatcherConsumerInstance == nil {
		workspaceWatcherConsumerOnce.Do(func() {
			workspaceWatcherConsumerInstance = webservice.NewWorkspaceWatcherConsumer()
		})
	}
	return workspaceWatcherConsumerInstance
}
