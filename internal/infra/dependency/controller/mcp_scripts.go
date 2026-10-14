package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpScriptsControllerOnce sync.Once
var mcpScriptsControllerInstance adapter.MCPToolController

func GetMCPScriptsController() adapter.MCPToolController {
	if mcpScriptsControllerInstance == nil {
		mcpScriptsControllerOnce.Do(func() {
			mcpScriptsControllerInstance = controller.NewMCPScripts(
				GetErrorHandler(),
				application.GetScriptToolsService(),
			)
		})
	}
	return mcpScriptsControllerInstance
}
