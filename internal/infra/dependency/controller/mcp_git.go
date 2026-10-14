package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpGitControllerOnce sync.Once
var mcpGitControllerInstance adapter.MCPToolController

func GetMCPGitController() adapter.MCPToolController {
	if mcpGitControllerInstance == nil {
		mcpGitControllerOnce.Do(func() {
			mcpGitControllerInstance = controller.NewMCPGit(
				GetErrorHandler(),
				application.GetManageSyncService(),
				application.GetResolveConflictsService(),
			)
		})
	}
	return mcpGitControllerInstance
}
