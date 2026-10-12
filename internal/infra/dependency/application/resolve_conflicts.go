package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var resolveConflictsServiceOnce sync.Once
var resolveConflictsServiceInstance adapter.ResolveConflictsService

func GetResolveConflictsService() adapter.ResolveConflictsService {
	if resolveConflictsServiceInstance == nil {
		resolveConflictsServiceOnce.Do(func() {
			resolveConflictsServiceInstance = service.NewResolveConflictsService(
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetGitAccountPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKnownHostsPersistence(),
				integration.GetGitWebservice(),
				integration.GetGitEventPublisher(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetTreePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetSyncSessionPersistence(),
				integration.GetSyncSessionPersistence(),
			)
		})
	}
	return resolveConflictsServiceInstance
}
