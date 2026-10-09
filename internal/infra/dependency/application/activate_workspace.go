package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var activateWorkspaceOnce sync.Once
var activateWorkspaceInstance adapter.ActivateWorkspaceService

func GetActivateWorkspaceService() adapter.ActivateWorkspaceService {
	if activateWorkspaceInstance == nil {
		activateWorkspaceOnce.Do(func() {
			activateWorkspaceInstance = service.NewActivateWorkspaceService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetWorkspacePersistence(),
			)
		})
	}
	return activateWorkspaceInstance
}
