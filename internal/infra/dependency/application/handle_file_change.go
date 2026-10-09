package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var handleFileChangeOnce sync.Once
var handleFileChangeInstance adapter.HandleFileChangeService

func GetHandleFileChangeService() adapter.HandleFileChangeService {
	if handleFileChangeInstance == nil {
		handleFileChangeOnce.Do(func() {
			handleFileChangeInstance = service.NewHandleFileChangeService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetWorkspaceWatcherWebservice(),
				integration.GetWorkspaceEventPublisher(),
			)
		})
	}
	return handleFileChangeInstance
}
