package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageCookiesOnce sync.Once
var manageCookiesInstance adapter.ManageCookiesService

func GetManageCookiesService() adapter.ManageCookiesService {
	if manageCookiesInstance == nil {
		manageCookiesOnce.Do(func() {
			manageCookiesInstance = service.NewManageCookiesService(
				integration.GetRegistryPersistence(),
				integration.GetCookieJarWebservice(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
				integration.GetWorkspaceEventPublisher(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKeychainWebservice(),
				integration.GetVariableEventPublisher(),
			)
		})
	}
	return manageCookiesInstance
}
