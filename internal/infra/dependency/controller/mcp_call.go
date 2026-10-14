package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpCallControllerOnce sync.Once
var mcpCallControllerInstance adapter.MCPCallController

func GetMCPCallController() adapter.MCPCallController {
	if mcpCallControllerInstance == nil {
		mcpCallControllerOnce.Do(func() {
			mcpCallControllerInstance = controller.NewMCPCall(
				GetErrorHandler(),
				application.GetListWorkspacesService(),
				application.GetGuardAIOutputService(),
				application.GetRecordAIActivityService(),
				application.GetNotifyAIChangeService(),
				application.GetApproveAISendService(),
			)
		})
	}
	return mcpCallControllerInstance
}
