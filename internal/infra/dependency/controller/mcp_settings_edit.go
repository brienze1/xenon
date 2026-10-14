package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpSettingsEditControllerOnce sync.Once
var mcpSettingsEditControllerInstance adapter.MCPToolController

func GetMCPSettingsEditController() adapter.MCPToolController {
	if mcpSettingsEditControllerInstance == nil {
		mcpSettingsEditControllerOnce.Do(func() {
			mcpSettingsEditControllerInstance = controller.NewMCPSettingsEdit(
				GetErrorHandler(),
				application.GetManageAppSettingsService(),
			)
		})
	}
	return mcpSettingsEditControllerInstance
}
