package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var saveRunnerPreferencesOnce sync.Once
var saveRunnerPreferencesInstance adapter.SaveRunnerPreferencesService

func GetSaveRunnerPreferencesService() adapter.SaveRunnerPreferencesService {
	if saveRunnerPreferencesInstance == nil {
		saveRunnerPreferencesOnce.Do(func() {
			saveRunnerPreferencesInstance = service.NewSaveRunnerPreferencesService(
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
			)
		})
	}
	return saveRunnerPreferencesInstance
}
