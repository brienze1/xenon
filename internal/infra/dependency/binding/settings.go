package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var settingsServiceOnce sync.Once
var settingsServiceInstance *binding.SettingsService

func GetSettingsService() *binding.SettingsService {
	if settingsServiceInstance == nil {
		settingsServiceOnce.Do(func() {
			settingsServiceInstance = binding.NewSettingsService(controller.GetSettingsController())
		})
	}
	return settingsServiceInstance
}
