package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageVariableTableOnce sync.Once
var manageVariableTableInstance adapter.ManageVariableTableService

func GetManageVariableTableService() adapter.ManageVariableTableService {
	if manageVariableTableInstance == nil {
		manageVariableTableOnce.Do(func() {
			manageVariableTableInstance = service.NewManageVariableTableService(
				integration.GetRegistryPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKeychainWebservice(),
				integration.GetVariableEventPublisher(),
			)
		})
	}
	return manageVariableTableInstance
}
