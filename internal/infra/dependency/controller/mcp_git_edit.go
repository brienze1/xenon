package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpGitEditControllerOnce sync.Once
var mcpGitEditControllerInstance adapter.MCPToolController

func GetMCPGitEditController() adapter.MCPToolController {
	if mcpGitEditControllerInstance == nil {
		mcpGitEditControllerOnce.Do(func() {
			mcpGitEditControllerInstance = controller.NewMCPGitEdit(
				GetErrorHandler(),
				application.GetManageSyncService(),
				application.GetLinkWorkspaceService(),
				application.GetCloneWorkspaceService(),
				application.GetCommitChangesService(),
				application.GetPullWorkspaceService(),
				application.GetResolveConflictsService(),
				application.GetFinishPullService(),
			)
		})
	}
	return mcpGitEditControllerInstance
}
