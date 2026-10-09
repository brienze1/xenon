package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceRemoveOnce sync.Once
var workspaceRemoveInstance adapter.WorkspaceRemoveController

func GetWorkspaceRemove() adapter.WorkspaceRemoveController {
	if workspaceRemoveInstance == nil {
		workspaceRemoveOnce.Do(func() {
			workspaceRemoveInstance = controller.NewWorkspaceRemove(GetErrorHandler(), application.GetRemoveWorkspaceService())
		})
	}
	return workspaceRemoveInstance
}
