package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var finishPullServiceOnce sync.Once
var finishPullServiceInstance adapter.FinishPullService

func GetFinishPullService() adapter.FinishPullService {
	if finishPullServiceInstance == nil {
		finishPullServiceOnce.Do(func() {
			finishPullServiceInstance = service.NewFinishPullService(
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
			)
		})
	}
	return finishPullServiceInstance
}
