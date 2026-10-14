package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpResponseControllerOnce sync.Once
var mcpResponseControllerInstance adapter.MCPToolController

func GetMCPResponseController() adapter.MCPToolController {
	if mcpResponseControllerInstance == nil {
		mcpResponseControllerOnce.Do(func() {
			mcpResponseControllerInstance = controller.NewMCPResponse(
				GetErrorHandler(),
				application.GetManageResponseService(),
				application.GetGuardAIOutputService(),
			)
		})
	}
	return mcpResponseControllerInstance
}
