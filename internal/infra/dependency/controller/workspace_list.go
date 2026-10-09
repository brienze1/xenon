package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceListControllerOnce sync.Once
var workspaceListControllerInstance adapter.WorkspaceListController

func GetWorkspaceListController() adapter.WorkspaceListController {
	if workspaceListControllerInstance == nil {
		workspaceListControllerOnce.Do(func() {
			workspaceListControllerInstance = controller.NewWorkspaceList(GetErrorHandler(), application.GetListWorkspacesService())
		})
	}
	return workspaceListControllerInstance
}
