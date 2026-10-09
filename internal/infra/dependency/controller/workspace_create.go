package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceCreateOnce sync.Once
var workspaceCreateInstance adapter.WorkspaceCreateController

func GetWorkspaceCreate() adapter.WorkspaceCreateController {
	if workspaceCreateInstance == nil {
		workspaceCreateOnce.Do(func() {
			workspaceCreateInstance = controller.NewWorkspaceCreate(GetErrorHandler(), application.GetCreateWorkspaceService())
		})
	}
	return workspaceCreateInstance
}
