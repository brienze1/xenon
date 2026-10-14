package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpExportControllerOnce sync.Once
var mcpExportControllerInstance adapter.MCPToolController

func GetMCPExportController() adapter.MCPToolController {
	if mcpExportControllerInstance == nil {
		mcpExportControllerOnce.Do(func() {
			mcpExportControllerInstance = controller.NewMCPExport(
				GetErrorHandler(),
				application.GetExportService(),
				application.GetManageResponseService(),
				application.GetGuardAIOutputService(),
				application.GetWriteAIOutputFileService(),
			)
		})
	}
	return mcpExportControllerInstance
}
