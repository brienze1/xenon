package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var setActiveEnvironmentOnce sync.Once
var setActiveEnvironmentInstance adapter.SetActiveEnvironmentService

func GetSetActiveEnvironmentService() adapter.SetActiveEnvironmentService {
	if setActiveEnvironmentInstance == nil {
		setActiveEnvironmentOnce.Do(func() {
			setActiveEnvironmentInstance = service.NewSetActiveEnvironmentService(
				integration.GetRegistryPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVariableEventPublisher(),
			)
		})
	}
	return setActiveEnvironmentInstance
}
