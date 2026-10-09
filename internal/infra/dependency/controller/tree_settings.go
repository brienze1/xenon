package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var treeSettingsOnce sync.Once
var treeSettingsInstance adapter.TreeSettingsController

func GetTreeSettings() adapter.TreeSettingsController {
	if treeSettingsInstance == nil {
		treeSettingsOnce.Do(func() {
			treeSettingsInstance = controller.NewTreeSettings(GetErrorHandler(), application.GetItemSettingsService())
		})
	}
	return treeSettingsInstance
}
