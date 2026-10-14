package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpWorkspaceEditControllerOnce sync.Once
var mcpWorkspaceEditControllerInstance adapter.MCPToolController

func GetMCPWorkspaceEditController() adapter.MCPToolController {
	if mcpWorkspaceEditControllerInstance == nil {
		mcpWorkspaceEditControllerOnce.Do(func() {
			mcpWorkspaceEditControllerInstance = controller.NewMCPWorkspaceEdit(
				GetErrorHandler(),
				application.GetCreateWorkspaceService(),
				application.GetOpenWorkspaceService(),
				application.GetRenameWorkspaceService(),
				application.GetActivateWorkspaceService(),
				application.GetAutoSyncService(),
			)
		})
	}
	return mcpWorkspaceEditControllerInstance
}
