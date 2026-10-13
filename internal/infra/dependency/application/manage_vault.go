package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageVaultOnce sync.Once
var manageVaultInstance adapter.ManageVaultService

func GetManageVaultService() adapter.ManageVaultService {
	if manageVaultInstance == nil {
		manageVaultOnce.Do(func() {
			manageVaultInstance = service.NewManageVaultService(
				integration.GetRegistryPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetTreePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetClipboardWebservice(),
				integration.GetVariableEventPublisher(),
			)
		})
	}
	return manageVaultInstance
}
