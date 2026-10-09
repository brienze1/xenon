package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceCreateControllerOnce sync.Once
var workspaceCreateControllerInstance adapter.WorkspaceCreateController

func GetWorkspaceCreateController() adapter.WorkspaceCreateController {
	if workspaceCreateControllerInstance == nil {
		workspaceCreateControllerOnce.Do(func() {
			workspaceCreateControllerInstance = controller.NewWorkspaceCreate(GetErrorHandler(), application.GetCreateWorkspaceService())
		})
	}
	return workspaceCreateControllerInstance
}
