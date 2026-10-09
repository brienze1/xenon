package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var workspaceWatcherWebserviceOnce sync.Once
var workspaceWatcherWebserviceInstance adapter.WorkspaceWatcherWebservice

func GetWorkspaceWatcherWebservice() adapter.WorkspaceWatcherWebservice {
	if workspaceWatcherWebserviceInstance == nil {
		workspaceWatcherWebserviceOnce.Do(func() {
			workspaceWatcherWebserviceInstance = webservice.NewWorkspaceWatcherWebservice(external.GetFileWatcher(), GetWorkspaceWatcherConsumer())
		})
	}
	return workspaceWatcherWebserviceInstance
}
