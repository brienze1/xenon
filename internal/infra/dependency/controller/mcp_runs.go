package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpRunsControllerOnce sync.Once
var mcpRunsControllerInstance adapter.MCPToolController

func GetMCPRunsController() adapter.MCPToolController {
	if mcpRunsControllerInstance == nil {
		mcpRunsControllerOnce.Do(func() {
			mcpRunsControllerInstance = controller.NewMCPRuns(GetErrorHandler(), application.GetSearchTreeService(), application.GetManageRunnerService())
		})
	}
	return mcpRunsControllerInstance
}
