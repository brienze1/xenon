package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceActivateControllerOnce sync.Once
var workspaceActivateControllerInstance adapter.WorkspaceActivateController

func GetWorkspaceActivateController() adapter.WorkspaceActivateController {
	if workspaceActivateControllerInstance == nil {
		workspaceActivateControllerOnce.Do(func() {
			workspaceActivateControllerInstance = controller.NewWorkspaceActivate(GetErrorHandler(), application.GetActivateWorkspaceService())
		})
	}
	return workspaceActivateControllerInstance
}
