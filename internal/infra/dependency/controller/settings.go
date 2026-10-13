package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var settingsControllerOnce sync.Once
var settingsControllerInstance adapter.SettingsController

func GetSettingsController() adapter.SettingsController {
	if settingsControllerInstance == nil {
		settingsControllerOnce.Do(func() {
			settingsControllerInstance = controller.NewSettings(GetErrorHandler(), application.GetManageAppSettingsService())
		})
	}
	return settingsControllerInstance
}
