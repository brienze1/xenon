package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var exportControllerOnce sync.Once
var exportControllerInstance adapter.ExportController

func GetExportController() adapter.ExportController {
	if exportControllerInstance == nil {
		exportControllerOnce.Do(func() {
			exportControllerInstance = controller.NewExport(GetErrorHandler(), application.GetExportService())
		})
	}
	return exportControllerInstance
}
