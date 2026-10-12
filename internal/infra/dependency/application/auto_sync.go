package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var autoSyncServiceOnce sync.Once
var autoSyncServiceInstance adapter.AutoSyncService

func GetAutoSyncService() adapter.AutoSyncService {
	if autoSyncServiceInstance == nil {
		autoSyncServiceOnce.Do(func() {
			autoSyncServiceInstance = service.NewAutoSyncService(
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetGitAccountPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKnownHostsPersistence(),
				integration.GetGitWebservice(),
				integration.GetGitEventPublisher(),
				integration.GetGitWebservice(),
				integration.GetSyncScheduler(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetTreePersistence(),
				integration.GetGitWebservice(),
				integration.GetTreePersistence(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetTreePersistence(),
				integration.GetWorkspaceEventPublisher(),
				integration.GetWorkspacePersistence(),
				integration.GetGitWebservice(),
				integration.GetSyncSessionPersistence(),
				integration.GetSyncSessionPersistence(),
				integration.GetTreePersistence(),
			)
		})
	}
	return autoSyncServiceInstance
}
