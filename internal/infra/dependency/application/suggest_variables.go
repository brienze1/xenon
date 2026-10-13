package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var suggestVariablesOnce sync.Once
var suggestVariablesInstance adapter.SuggestVariablesService

func GetSuggestVariablesService() adapter.SuggestVariablesService {
	if suggestVariablesInstance == nil {
		suggestVariablesOnce.Do(func() {
			suggestVariablesInstance = service.NewSuggestVariablesService(
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
	return suggestVariablesInstance
}
