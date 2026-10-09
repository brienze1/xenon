package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var workspaceOpenOnce sync.Once
var workspaceOpenInstance adapter.WorkspaceOpenController

func GetWorkspaceOpen() adapter.WorkspaceOpenController {
	if workspaceOpenInstance == nil {
		workspaceOpenOnce.Do(func() {
			workspaceOpenInstance = controller.NewWorkspaceOpen(GetErrorHandler(), application.GetOpenWorkspaceService())
		})
	}
	return workspaceOpenInstance
}
