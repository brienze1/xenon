package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var itemSettingsOnce sync.Once
var itemSettingsInstance adapter.ItemSettingsService

func GetItemSettingsService() adapter.ItemSettingsService {
	if itemSettingsInstance == nil {
		itemSettingsOnce.Do(func() {
			itemSettingsInstance = service.NewItemSettingsService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
			)
		})
	}
	return itemSettingsInstance
}
