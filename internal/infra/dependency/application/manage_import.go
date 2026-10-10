package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageImportOnce sync.Once
var manageImportInstance adapter.ManageImportService

func GetManageImportService() adapter.ManageImportService {
	if manageImportInstance == nil {
		manageImportOnce.Do(func() {
			manageImportInstance = service.NewManageImportService(
				integration.GetImportSourcePersistence(),
				integration.GetImportSourcePersistence(),
				integration.GetImportSourcePersistence(),
				integration.GetImportSourcePersistence(),
				integration.GetPostmanWebservice(),
				integration.GetDialogWebservice(),
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetImportStagingPersistence(),
				integration.GetImportStagingPersistence(),
				integration.GetImportStagingPersistence(),
				integration.GetImportProgressPublisher(),
				integration.GetWorkspaceEventPublisher(),
			)
		})
	}
	return manageImportInstance
}
