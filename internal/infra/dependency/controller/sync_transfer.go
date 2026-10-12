package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var syncTransferControllerOnce sync.Once
var syncTransferControllerInstance adapter.SyncTransferController

func GetSyncTransferController() adapter.SyncTransferController {
	if syncTransferControllerInstance == nil {
		syncTransferControllerOnce.Do(func() {
			syncTransferControllerInstance = controller.NewSyncTransfer(GetErrorHandler(), application.GetCommitChangesService(), application.GetPullWorkspaceService(), application.GetPushWorkspaceService())
		})
	}
	return syncTransferControllerInstance
}
