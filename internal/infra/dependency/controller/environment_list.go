package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var environmentListControllerOnce sync.Once
var environmentListControllerInstance adapter.EnvironmentListController

func GetEnvironmentListController() adapter.EnvironmentListController {
	if environmentListControllerInstance == nil {
		environmentListControllerOnce.Do(func() {
			environmentListControllerInstance = controller.NewEnvironmentList(GetErrorHandler(), application.GetManageEnvironmentService())
		})
	}
	return environmentListControllerInstance
}
