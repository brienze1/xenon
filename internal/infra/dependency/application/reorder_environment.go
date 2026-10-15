package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var reorderEnvironmentOnce sync.Once
var reorderEnvironmentInstance adapter.ReorderEnvironmentService

func GetReorderEnvironmentService() adapter.ReorderEnvironmentService {
	if reorderEnvironmentInstance == nil {
		reorderEnvironmentOnce.Do(func() {
			reorderEnvironmentInstance = service.NewReorderEnvironmentService(
				integration.GetRegistryPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVariableEventPublisher(),
			)
		})
	}
	return reorderEnvironmentInstance
}
