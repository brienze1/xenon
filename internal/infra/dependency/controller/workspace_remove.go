package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceRemoveControllerOnce sync.Once
var workspaceRemoveControllerInstance adapter.WorkspaceRemoveController

func GetWorkspaceRemoveController() adapter.WorkspaceRemoveController {
	if workspaceRemoveControllerInstance == nil {
		workspaceRemoveControllerOnce.Do(func() {
			workspaceRemoveControllerInstance = controller.NewWorkspaceRemove(GetErrorHandler(), application.GetRemoveWorkspaceService())
		})
	}
	return workspaceRemoveControllerInstance
}
