package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceListOnce sync.Once
var workspaceListInstance adapter.WorkspaceListController

func GetWorkspaceList() adapter.WorkspaceListController {
	if workspaceListInstance == nil {
		workspaceListOnce.Do(func() {
			workspaceListInstance = controller.NewWorkspaceList(GetErrorHandler(), application.GetListWorkspacesService())
		})
	}
	return workspaceListInstance
}
