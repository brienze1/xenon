package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpGitPushControllerOnce sync.Once
var mcpGitPushControllerInstance adapter.MCPToolController

func GetMCPGitPushController() adapter.MCPToolController {
	if mcpGitPushControllerInstance == nil {
		mcpGitPushControllerOnce.Do(func() {
			mcpGitPushControllerInstance = controller.NewMCPGitPush(
				GetErrorHandler(),
				application.GetPushWorkspaceService(),
			)
		})
	}
	return mcpGitPushControllerInstance
}
