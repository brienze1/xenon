package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpEnvEditControllerOnce sync.Once
var mcpEnvEditControllerInstance adapter.MCPToolController

func GetMCPEnvEditController() adapter.MCPToolController {
	if mcpEnvEditControllerInstance == nil {
		mcpEnvEditControllerOnce.Do(func() {
			mcpEnvEditControllerInstance = controller.NewMCPEnvEdit(
				GetErrorHandler(),
				application.GetManageEnvironmentService(),
				application.GetReorderEnvironmentService(),
				application.GetAutoSyncService(),
				application.GetSetActiveEnvironmentService(),
				application.GetManageVariableTableService(),
			)
		})
	}
	return mcpEnvEditControllerInstance
}
