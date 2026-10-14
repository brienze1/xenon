package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpHistoryControllerOnce sync.Once
var mcpHistoryControllerInstance adapter.MCPToolController

func GetMCPHistoryController() adapter.MCPToolController {
	if mcpHistoryControllerInstance == nil {
		mcpHistoryControllerOnce.Do(func() {
			mcpHistoryControllerInstance = controller.NewMCPHistory(
				GetErrorHandler(),
				application.GetManageHistoryService(),
			)
		})
	}
	return mcpHistoryControllerInstance
}
