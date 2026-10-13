package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var renameWorkspaceOnce sync.Once
var renameWorkspaceInstance adapter.RenameWorkspaceService

func GetRenameWorkspaceService() adapter.RenameWorkspaceService {
	if renameWorkspaceInstance == nil {
		renameWorkspaceOnce.Do(func() {
			renameWorkspaceInstance = service.NewRenameWorkspaceService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
			)
		})
	}
	return renameWorkspaceInstance
}
