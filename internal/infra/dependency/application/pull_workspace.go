package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var pullWorkspaceServiceOnce sync.Once
var pullWorkspaceServiceInstance adapter.PullWorkspaceService

func GetPullWorkspaceService() adapter.PullWorkspaceService {
	if pullWorkspaceServiceInstance == nil {
		pullWorkspaceServiceOnce.Do(func() {
			pullWorkspaceServiceInstance = service.NewPullWorkspaceService(
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
				integration.GetSyncSessionPersistence(),
				integration.GetSyncSessionPersistence(),
				integration.GetTreePersistence(),
			)
		})
	}
	return pullWorkspaceServiceInstance
}
