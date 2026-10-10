package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var inspectVariablesOnce sync.Once
var inspectVariablesInstance adapter.InspectVariablesService

func GetInspectVariablesService() adapter.InspectVariablesService {
	if inspectVariablesInstance == nil {
		inspectVariablesOnce.Do(func() {
			inspectVariablesInstance = service.NewInspectVariablesService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVaultPersistence(),
				integration.GetDynamicValuesWebservice(),
			)
		})
	}
	return inspectVariablesInstance
}
