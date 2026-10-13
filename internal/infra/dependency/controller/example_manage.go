package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var exampleManageControllerOnce sync.Once
var exampleManageControllerInstance adapter.ExampleManageController

func GetExampleManageController() adapter.ExampleManageController {
	if exampleManageControllerInstance == nil {
		exampleManageControllerOnce.Do(func() {
			exampleManageControllerInstance = controller.NewExampleManage(GetErrorHandler(), application.GetManageExampleService())
		})
	}
	return exampleManageControllerInstance
}
