package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var createWorkspaceOnce sync.Once
var createWorkspaceInstance adapter.CreateWorkspaceService

func GetCreateWorkspaceService() adapter.CreateWorkspaceService {
	if createWorkspaceInstance == nil {
		createWorkspaceOnce.Do(func() {
			createWorkspaceInstance = service.NewCreateWorkspaceService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
			)
		})
	}
	return createWorkspaceInstance
}
