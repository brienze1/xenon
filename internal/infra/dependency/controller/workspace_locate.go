package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceLocateOnce sync.Once
var workspaceLocateInstance adapter.WorkspaceLocateController

func GetWorkspaceLocate() adapter.WorkspaceLocateController {
	if workspaceLocateInstance == nil {
		workspaceLocateOnce.Do(func() {
			workspaceLocateInstance = controller.NewWorkspaceLocate(GetErrorHandler(), application.GetLocateWorkspaceService())
		})
	}
	return workspaceLocateInstance
}
