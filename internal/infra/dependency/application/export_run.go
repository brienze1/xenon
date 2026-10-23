package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var exportRunOnce sync.Once
var exportRunInstance adapter.ExportRunService

func GetExportRunService() adapter.ExportRunService {
	if exportRunInstance == nil {
		exportRunOnce.Do(func() {
			exportRunInstance = service.NewExportRunService(
				integration.GetActiveRunsPersistence(),
				integration.GetRunPersistence(),
				integration.GetExportFilePersistence(),
			)
		})
	}
	return exportRunInstance
}
