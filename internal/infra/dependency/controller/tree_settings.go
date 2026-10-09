package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var treeSettingsControllerOnce sync.Once
var treeSettingsControllerInstance adapter.TreeSettingsController

func GetTreeSettingsController() adapter.TreeSettingsController {
	if treeSettingsControllerInstance == nil {
		treeSettingsControllerOnce.Do(func() {
			treeSettingsControllerInstance = controller.NewTreeSettings(GetErrorHandler(), application.GetManageItemSettingsService())
		})
	}
	return treeSettingsControllerInstance
}
