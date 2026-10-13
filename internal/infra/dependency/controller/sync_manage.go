package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var syncManageControllerOnce sync.Once
var syncManageControllerInstance adapter.SyncManageController

func GetSyncManageController() adapter.SyncManageController {
	if syncManageControllerInstance == nil {
		syncManageControllerOnce.Do(func() {
			syncManageControllerInstance = controller.NewSyncManage(GetErrorHandler(), application.GetManageSyncService())
		})
	}
	return syncManageControllerInstance
}
