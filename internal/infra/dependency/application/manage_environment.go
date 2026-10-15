package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageEnvironmentOnce sync.Once
var manageEnvironmentInstance adapter.ManageEnvironmentService

func GetManageEnvironmentService() adapter.ManageEnvironmentService {
	if manageEnvironmentInstance == nil {
		manageEnvironmentOnce.Do(func() {
			manageEnvironmentInstance = service.NewManageEnvironmentService(
				integration.GetRegistryPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVariableEventPublisher(),
			)
		})
	}
	return manageEnvironmentInstance
}
