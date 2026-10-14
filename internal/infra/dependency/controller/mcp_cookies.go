package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpCookiesControllerOnce sync.Once
var mcpCookiesControllerInstance adapter.MCPToolController

func GetMCPCookiesController() adapter.MCPToolController {
	if mcpCookiesControllerInstance == nil {
		mcpCookiesControllerOnce.Do(func() {
			mcpCookiesControllerInstance = controller.NewMCPCookies(
				GetErrorHandler(),
				application.GetManageCookiesService(),
			)
		})
	}
	return mcpCookiesControllerInstance
}
