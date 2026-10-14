package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpSettingsControllerOnce sync.Once
var mcpSettingsControllerInstance adapter.MCPToolController

func GetMCPSettingsController() adapter.MCPToolController {
	if mcpSettingsControllerInstance == nil {
		mcpSettingsControllerOnce.Do(func() {
			mcpSettingsControllerInstance = controller.NewMCPSettings(
				GetErrorHandler(),
				application.GetManageAppSettingsService(),
			)
		})
	}
	return mcpSettingsControllerInstance
}
