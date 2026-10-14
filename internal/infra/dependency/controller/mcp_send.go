package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpSendControllerOnce sync.Once
var mcpSendControllerInstance adapter.MCPToolController

func GetMCPSendController() adapter.MCPToolController {
	if mcpSendControllerInstance == nil {
		mcpSendControllerOnce.Do(func() {
			mcpSendControllerInstance = controller.NewMCPSend(
				GetErrorHandler(),
				application.GetSearchTreeService(),
				application.GetSaveRequestService(),
				application.GetSendRequestService(),
				application.GetManageAppSettingsService(),
				application.GetApproveAISendService(),
				application.GetManageResponseService(),
				application.GetGuardAIOutputService(),
			)
		})
	}
	return mcpSendControllerInstance
}
