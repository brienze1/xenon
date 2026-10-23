package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var removeWorkspaceOnce sync.Once
var removeWorkspaceInstance adapter.RemoveWorkspaceService

func GetRemoveWorkspaceService() adapter.RemoveWorkspaceService {
	if removeWorkspaceInstance == nil {
		removeWorkspaceOnce.Do(func() {
			removeWorkspaceInstance = service.NewRemoveWorkspaceService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetTrashWebservice(),
				integration.GetUIStatePersistence(),
				integration.GetWorkspaceWatcherWebservice(),
				integration.GetVaultPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetRunPersistence(),
			)
		})
	}
	return removeWorkspaceInstance
}
