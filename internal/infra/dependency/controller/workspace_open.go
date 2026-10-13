package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceOpenControllerOnce sync.Once
var workspaceOpenControllerInstance adapter.WorkspaceOpenController

func GetWorkspaceOpenController() adapter.WorkspaceOpenController {
	if workspaceOpenControllerInstance == nil {
		workspaceOpenControllerOnce.Do(func() {
			workspaceOpenControllerInstance = controller.NewWorkspaceOpen(GetErrorHandler(), application.GetOpenWorkspaceService())
		})
	}
	return workspaceOpenControllerInstance
}
