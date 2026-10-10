package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var environmentSetActiveControllerOnce sync.Once
var environmentSetActiveControllerInstance adapter.EnvironmentSetActiveController

func GetEnvironmentSetActiveController() adapter.EnvironmentSetActiveController {
	if environmentSetActiveControllerInstance == nil {
		environmentSetActiveControllerOnce.Do(func() {
			environmentSetActiveControllerInstance = controller.NewEnvironmentSetActive(GetErrorHandler(), application.GetManageEnvironmentService())
		})
	}
	return environmentSetActiveControllerInstance
}
