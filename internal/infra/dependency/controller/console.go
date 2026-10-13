package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var consoleControllerOnce sync.Once
var consoleControllerInstance adapter.ConsoleController

func GetConsoleController() adapter.ConsoleController {
	if consoleControllerInstance == nil {
		consoleControllerOnce.Do(func() {
			consoleControllerInstance = controller.NewConsole(GetErrorHandler(), application.GetManageConsoleService())
		})
	}
	return consoleControllerInstance
}
