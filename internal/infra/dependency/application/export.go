package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var exportOnce sync.Once
var exportInstance adapter.ExportService

func GetExportService() adapter.ExportService {
	if exportInstance == nil {
		exportOnce.Do(func() {
			exportInstance = service.NewExportService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetRequestPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetPostmanWebservice(),
				integration.GetXenonExportPersistence(),
				integration.GetXenonExportPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetExportFilePersistence(),
				integration.GetDialogWebservice(),
				integration.GetFileManagerWebservice(),
				integration.GetExportProgressPublisher(),
			)
		})
	}
	return exportInstance
}
