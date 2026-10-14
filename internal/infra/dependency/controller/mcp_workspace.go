package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpWorkspaceControllerOnce sync.Once
var mcpWorkspaceControllerInstance adapter.MCPToolController

func GetMCPWorkspaceController() adapter.MCPToolController {
	if mcpWorkspaceControllerInstance == nil {
		mcpWorkspaceControllerOnce.Do(func() {
			mcpWorkspaceControllerInstance = controller.NewMCPWorkspace(
				GetErrorHandler(),
				application.GetListWorkspacesService(),
				application.GetManageEnvironmentService(),
			)
		})
	}
	return mcpWorkspaceControllerInstance
}
