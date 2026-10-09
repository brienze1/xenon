package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceRenameOnce sync.Once
var workspaceRenameInstance adapter.WorkspaceRenameController

func GetWorkspaceRename() adapter.WorkspaceRenameController {
	if workspaceRenameInstance == nil {
		workspaceRenameOnce.Do(func() {
			workspaceRenameInstance = controller.NewWorkspaceRename(GetErrorHandler(), application.GetRenameWorkspaceService())
		})
	}
	return workspaceRenameInstance
}
