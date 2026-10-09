package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var locateWorkspaceOnce sync.Once
var locateWorkspaceInstance adapter.LocateWorkspaceService

func GetLocateWorkspaceService() adapter.LocateWorkspaceService {
	if locateWorkspaceInstance == nil {
		locateWorkspaceOnce.Do(func() {
			locateWorkspaceInstance = service.NewLocateWorkspaceService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
			)
		})
	}
	return locateWorkspaceInstance
}
