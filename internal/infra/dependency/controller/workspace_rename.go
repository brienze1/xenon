package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceRenameControllerOnce sync.Once
var workspaceRenameControllerInstance adapter.WorkspaceRenameController

func GetWorkspaceRenameController() adapter.WorkspaceRenameController {
	if workspaceRenameControllerInstance == nil {
		workspaceRenameControllerOnce.Do(func() {
			workspaceRenameControllerInstance = controller.NewWorkspaceRename(GetErrorHandler(), application.GetRenameWorkspaceService())
		})
	}
	return workspaceRenameControllerInstance
}
