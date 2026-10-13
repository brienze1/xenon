package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var syncLinkControllerOnce sync.Once
var syncLinkControllerInstance adapter.SyncLinkController

func GetSyncLinkController() adapter.SyncLinkController {
	if syncLinkControllerInstance == nil {
		syncLinkControllerOnce.Do(func() {
			syncLinkControllerInstance = controller.NewSyncLink(GetErrorHandler(), application.GetLinkWorkspaceService(), application.GetCloneWorkspaceService(), application.GetUnlinkWorkspaceService())
		})
	}
	return syncLinkControllerInstance
}
