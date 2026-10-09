package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var windowStateOnce sync.Once
var windowStateInstance adapter.WindowStateController

func GetWindowState() adapter.WindowStateController {
	if windowStateInstance == nil {
		windowStateOnce.Do(func() {
			windowStateInstance = controller.NewWindowState(GetErrorHandler(), application.GetWindowStateService())
		})
	}
	return windowStateInstance
}
