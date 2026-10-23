package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpTreeControllerOnce sync.Once
var mcpTreeControllerInstance adapter.MCPToolController

func GetMCPTreeController() adapter.MCPToolController {
	if mcpTreeControllerInstance == nil {
		mcpTreeControllerOnce.Do(func() {
			mcpTreeControllerInstance = controller.NewMCPTree(
				GetErrorHandler(),
				application.GetSearchTreeService(),
				application.GetSaveRequestService(),
				application.GetManageItemSettingsService(),
				application.GetManageExampleService(),
				application.GetManageVariableTableService(),
				application.GetManageOAuthService(),
				application.GetManageEnvironmentService(),
			)
		})
	}
	return mcpTreeControllerInstance
}
