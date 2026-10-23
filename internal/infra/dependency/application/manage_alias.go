package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageAliasOnce sync.Once
var manageAliasInstance adapter.ManageAliasService

func GetManageAliasService() adapter.ManageAliasService {
	if manageAliasInstance == nil {
		manageAliasOnce.Do(func() {
			manageAliasInstance = service.NewManageAliasService(
				integration.GetRegistryPersistence(),
				integration.GetRequestPersistence(),
				integration.GetRequestPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKeychainWebservice(),
				integration.GetVariableEventPublisher(),
				integration.GetEnvironmentPersistence(),
			)
		})
	}
	return manageAliasInstance
}
