package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var uiStateOnce sync.Once
var uiStateInstance adapter.UIStateService

func GetUIStateService() adapter.UIStateService {
	if uiStateInstance == nil {
		uiStateOnce.Do(func() {
			uiStateInstance = service.NewUIStateService(
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
			)
		})
	}
	return uiStateInstance
}
