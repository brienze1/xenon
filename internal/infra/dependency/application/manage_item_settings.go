package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageItemSettingsOnce sync.Once
var manageItemSettingsInstance adapter.ManageItemSettingsService

func GetManageItemSettingsService() adapter.ManageItemSettingsService {
	if manageItemSettingsInstance == nil {
		manageItemSettingsOnce.Do(func() {
			manageItemSettingsInstance = service.NewManageItemSettingsService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
			)
		})
	}
	return manageItemSettingsInstance
}
