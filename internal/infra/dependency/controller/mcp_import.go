package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpImportControllerOnce sync.Once
var mcpImportControllerInstance adapter.MCPToolController

func GetMCPImportController() adapter.MCPToolController {
	if mcpImportControllerInstance == nil {
		mcpImportControllerOnce.Do(func() {
			mcpImportControllerInstance = controller.NewMCPImport(
				GetErrorHandler(),
				application.GetManageImportService(),
			)
		})
	}
	return mcpImportControllerInstance
}
