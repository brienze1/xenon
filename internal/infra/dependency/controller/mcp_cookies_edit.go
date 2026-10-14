package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpCookiesEditControllerOnce sync.Once
var mcpCookiesEditControllerInstance adapter.MCPToolController

func GetMCPCookiesEditController() adapter.MCPToolController {
	if mcpCookiesEditControllerInstance == nil {
		mcpCookiesEditControllerOnce.Do(func() {
			mcpCookiesEditControllerInstance = controller.NewMCPCookiesEdit(
				GetErrorHandler(),
				application.GetManageCookiesService(),
			)
		})
	}
	return mcpCookiesEditControllerInstance
}
