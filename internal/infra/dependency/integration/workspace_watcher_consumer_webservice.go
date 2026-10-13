package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var workspaceWatcherConsumerWebserviceOnce sync.Once
var workspaceWatcherConsumerWebserviceInstance adapter.WorkspaceWatcherConsumerWebservice

func GetWorkspaceWatcherConsumerWebservice() adapter.WorkspaceWatcherConsumerWebservice {
	if workspaceWatcherConsumerWebserviceInstance == nil {
		workspaceWatcherConsumerWebserviceOnce.Do(func() {
			workspaceWatcherConsumerWebserviceInstance = webservice.NewWorkspaceWatcherConsumerWebservice()
		})
	}
	return workspaceWatcherConsumerWebserviceInstance
}
