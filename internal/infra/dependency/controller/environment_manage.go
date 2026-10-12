package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var environmentManageControllerOnce sync.Once
var environmentManageControllerInstance adapter.EnvironmentManageController

func GetEnvironmentManageController() adapter.EnvironmentManageController {
	if environmentManageControllerInstance == nil {
		environmentManageControllerOnce.Do(func() {
			environmentManageControllerInstance = controller.NewEnvironmentManage(GetErrorHandler(), application.GetManageEnvironmentService(), application.GetAutoSyncService())
		})
	}
	return environmentManageControllerInstance
}
