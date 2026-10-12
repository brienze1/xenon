package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var linkWorkspaceServiceOnce sync.Once
var linkWorkspaceServiceInstance adapter.LinkWorkspaceService

func GetLinkWorkspaceService() adapter.LinkWorkspaceService {
	if linkWorkspaceServiceInstance == nil {
		linkWorkspaceServiceOnce.Do(func() {
			linkWorkspaceServiceInstance = service.NewLinkWorkspaceService(
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
				integration.GetTreePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspaceWatcherWebservice(),
			)
		})
	}
	return linkWorkspaceServiceInstance
}
