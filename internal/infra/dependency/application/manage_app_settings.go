package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageAppSettingsOnce sync.Once
var manageAppSettingsInstance adapter.ManageAppSettingsService

func GetManageAppSettingsService() adapter.ManageAppSettingsService {
	if manageAppSettingsInstance == nil {
		manageAppSettingsOnce.Do(func() {
			manageAppSettingsInstance = service.NewManageAppSettingsService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetAIAccessWebservice(),
				integration.GetAIAccessStatusPersistence(),
				integration.GetAIEventPublisher(),
			)
		})
	}
	return manageAppSettingsInstance
}
