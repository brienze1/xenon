package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var cloneWorkspaceServiceOnce sync.Once
var cloneWorkspaceServiceInstance adapter.CloneWorkspaceService

func GetCloneWorkspaceService() adapter.CloneWorkspaceService {
	if cloneWorkspaceServiceInstance == nil {
		cloneWorkspaceServiceOnce.Do(func() {
			cloneWorkspaceServiceInstance = service.NewCloneWorkspaceService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
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
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
			)
		})
	}
	return cloneWorkspaceServiceInstance
}
