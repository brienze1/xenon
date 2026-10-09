package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var windowStateControllerOnce sync.Once
var windowStateControllerInstance adapter.WindowStateController

func GetWindowStateController() adapter.WindowStateController {
	if windowStateControllerInstance == nil {
		windowStateControllerOnce.Do(func() {
			windowStateControllerInstance = controller.NewWindowState(GetErrorHandler(), application.GetManageWindowStateService())
		})
	}
	return windowStateControllerInstance
}
