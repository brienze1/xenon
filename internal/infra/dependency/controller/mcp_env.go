package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpEnvControllerOnce sync.Once
var mcpEnvControllerInstance adapter.MCPToolController

func GetMCPEnvController() adapter.MCPToolController {
	if mcpEnvControllerInstance == nil {
		mcpEnvControllerOnce.Do(func() {
			mcpEnvControllerInstance = controller.NewMCPEnv(
				GetErrorHandler(),
				application.GetManageEnvironmentService(),
				application.GetManageVariableTableService(),
				application.GetSearchTreeService(),
				application.GetSaveRequestService(),
				application.GetInspectVariablesService(),
			)
		})
	}
	return mcpEnvControllerInstance
}
