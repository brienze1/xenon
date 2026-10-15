package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var updateControllerOnce sync.Once
var updateControllerInstance adapter.UpdateController

func GetUpdateController() adapter.UpdateController {
	if updateControllerInstance == nil {
		updateControllerOnce.Do(func() {
			updateControllerInstance = controller.NewUpdate(GetErrorHandler(), application.GetManageUpdateService())
		})
	}
	return updateControllerInstance
}
