package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpTreeEditControllerOnce sync.Once
var mcpTreeEditControllerInstance adapter.MCPToolController

func GetMCPTreeEditController() adapter.MCPToolController {
	if mcpTreeEditControllerInstance == nil {
		mcpTreeEditControllerOnce.Do(func() {
			mcpTreeEditControllerInstance = controller.NewMCPTreeEdit(
				GetErrorHandler(),
				application.GetSearchTreeService(),
				application.GetManageTreeItemService(),
				application.GetSaveRequestService(),
				application.GetAutoSyncService(),
				application.GetManageExampleService(),
				application.GetManageAliasService(),
				application.GetManageResponseService(),
				application.GetGuardAIOutputService(),
				application.GetManageItemSettingsService(),
				application.GetManageVariableTableService(),
				application.GetManageOAuthService(),
				application.GetManageEnvironmentService(),
			)
		})
	}
	return mcpTreeEditControllerInstance
}
