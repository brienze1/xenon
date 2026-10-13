package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var openWorkspaceOnce sync.Once
var openWorkspaceInstance adapter.OpenWorkspaceService

func GetOpenWorkspaceService() adapter.OpenWorkspaceService {
	if openWorkspaceInstance == nil {
		openWorkspaceOnce.Do(func() {
			openWorkspaceInstance = service.NewOpenWorkspaceService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
			)
		})
	}
	return openWorkspaceInstance
}
