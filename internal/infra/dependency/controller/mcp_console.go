package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpConsoleControllerOnce sync.Once
var mcpConsoleControllerInstance adapter.MCPToolController

func GetMCPConsoleController() adapter.MCPToolController {
	if mcpConsoleControllerInstance == nil {
		mcpConsoleControllerOnce.Do(func() {
			mcpConsoleControllerInstance = controller.NewMCPConsole(
				GetErrorHandler(),
				application.GetManageConsoleService(),
			)
		})
	}
	return mcpConsoleControllerInstance
}
