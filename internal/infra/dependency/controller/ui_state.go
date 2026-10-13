package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var uiStateControllerOnce sync.Once
var uiStateControllerInstance adapter.UIStateController

func GetUIStateController() adapter.UIStateController {
	if uiStateControllerInstance == nil {
		uiStateControllerOnce.Do(func() {
			uiStateControllerInstance = controller.NewUIState(GetErrorHandler(), application.GetManageUIStateService())
		})
	}
	return uiStateControllerInstance
}
