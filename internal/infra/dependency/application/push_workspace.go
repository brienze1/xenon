package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var pushWorkspaceServiceOnce sync.Once
var pushWorkspaceServiceInstance adapter.PushWorkspaceService

func GetPushWorkspaceService() adapter.PushWorkspaceService {
	if pushWorkspaceServiceInstance == nil {
		pushWorkspaceServiceOnce.Do(func() {
			pushWorkspaceServiceInstance = service.NewPushWorkspaceService(
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
				integration.GetGitWebservice(),
			)
		})
	}
	return pushWorkspaceServiceInstance
}
