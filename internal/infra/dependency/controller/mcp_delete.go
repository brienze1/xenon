package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpDeleteControllerOnce sync.Once
var mcpDeleteControllerInstance adapter.MCPToolController

func GetMCPDeleteController() adapter.MCPToolController {
	if mcpDeleteControllerInstance == nil {
		mcpDeleteControllerOnce.Do(func() {
			mcpDeleteControllerInstance = controller.NewMCPDelete(
				GetErrorHandler(),
				application.GetSearchTreeService(),
				application.GetManageTreeItemService(),
				application.GetManageExampleService(),
				application.GetManageEnvironmentService(),
				application.GetManageCookiesService(),
				application.GetRemoveWorkspaceService(),
				application.GetManageOAuthService(),
				application.GetUnlinkWorkspaceService(),
				application.GetManageHistoryService(),
				application.GetManageConsoleService(),
			)
		})
	}
	return mcpDeleteControllerInstance
}
