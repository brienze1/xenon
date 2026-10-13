package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var unlinkWorkspaceServiceOnce sync.Once
var unlinkWorkspaceServiceInstance adapter.UnlinkWorkspaceService

func GetUnlinkWorkspaceService() adapter.UnlinkWorkspaceService {
	if unlinkWorkspaceServiceInstance == nil {
		unlinkWorkspaceServiceOnce.Do(func() {
			unlinkWorkspaceServiceInstance = service.NewUnlinkWorkspaceService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
				integration.GetGitWebservice(),
				integration.GetSyncSessionPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspaceWatcherWebservice(),
			)
		})
	}
	return unlinkWorkspaceServiceInstance
}
