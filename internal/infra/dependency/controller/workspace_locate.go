package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceLocateControllerOnce sync.Once
var workspaceLocateControllerInstance adapter.WorkspaceLocateController

func GetWorkspaceLocateController() adapter.WorkspaceLocateController {
	if workspaceLocateControllerInstance == nil {
		workspaceLocateControllerOnce.Do(func() {
			workspaceLocateControllerInstance = controller.NewWorkspaceLocate(GetErrorHandler(), application.GetLocateWorkspaceService())
		})
	}
	return workspaceLocateControllerInstance
}
