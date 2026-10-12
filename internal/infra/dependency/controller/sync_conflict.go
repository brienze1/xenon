package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var syncConflictControllerOnce sync.Once
var syncConflictControllerInstance adapter.SyncConflictController

func GetSyncConflictController() adapter.SyncConflictController {
	if syncConflictControllerInstance == nil {
		syncConflictControllerOnce.Do(func() {
			syncConflictControllerInstance = controller.NewSyncConflict(GetErrorHandler(), application.GetResolveConflictsService(), application.GetFinishPullService())
		})
	}
	return syncConflictControllerInstance
}
