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
				integration.GetPostmanWebservice(),
				integration.GetPostmanWebservice(),
				integration.GetBrunoWebservice(),
				integration.GetCurlWebservice(),
				integration.GetXenonExportPersistence(),
				integration.GetXenonExportPersistence(),
				integration.GetDialogWebservice(),
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetRegistryPersistence(),
				integration.GetImportStagingPersistence(),
				integration.GetImportStagingPersistence(),
				integration.GetImportStagingPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKeychainWebservice(),
				integration.GetImportProgressPublisher(),
				integration.GetWorkspaceEventPublisher(),
				integration.GetVariableEventPublisher(),
			)
		})
	}
	return manageImportInstance
}
