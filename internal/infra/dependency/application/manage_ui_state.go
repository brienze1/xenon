package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageUIStateOnce sync.Once
var manageUIStateInstance adapter.ManageUIStateService

func GetManageUIStateService() adapter.ManageUIStateService {
	if manageUIStateInstance == nil {
		manageUIStateOnce.Do(func() {
			manageUIStateInstance = service.NewManageUIStateService(
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
			)
		})
	}
	return manageUIStateInstance
}
