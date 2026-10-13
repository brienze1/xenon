package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageSyncServiceOnce sync.Once
var manageSyncServiceInstance adapter.ManageSyncService

func GetManageSyncService() adapter.ManageSyncService {
	if manageSyncServiceInstance == nil {
		manageSyncServiceOnce.Do(func() {
			manageSyncServiceInstance = service.NewManageSyncService(
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
				integration.GetGitAccountPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKnownHostsPersistence(),
				integration.GetGitWebservice(),
				integration.GetGitEventPublisher(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetTreePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetSyncSessionPersistence(),
			)
		})
	}
	return manageSyncServiceInstance
}
