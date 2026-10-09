package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceActivateOnce sync.Once
var workspaceActivateInstance adapter.WorkspaceActivateController

func GetWorkspaceActivate() adapter.WorkspaceActivateController {
	if workspaceActivateInstance == nil {
		workspaceActivateOnce.Do(func() {
			workspaceActivateInstance = controller.NewWorkspaceActivate(GetErrorHandler(), application.GetActivateWorkspaceService())
		})
	}
	return workspaceActivateInstance
}
