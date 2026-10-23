package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpRunControllerOnce sync.Once
var mcpRunControllerInstance adapter.MCPToolController

func GetMCPRunController() adapter.MCPToolController {
	if mcpRunControllerInstance == nil {
		mcpRunControllerOnce.Do(func() {
			mcpRunControllerInstance = controller.NewMCPRun(
				GetErrorHandler(),
				application.GetSearchTreeService(),
				application.GetManageRunnerService(),
				application.GetManageAppSettingsService(),
				application.GetApproveAISendService(),
			)
		})
	}
	return mcpRunControllerInstance
}
