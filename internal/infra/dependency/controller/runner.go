package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var runnerControllerOnce sync.Once
var runnerControllerInstance adapter.RunnerController

func GetRunnerController() adapter.RunnerController {
	if runnerControllerInstance == nil {
		runnerControllerOnce.Do(func() {
			runnerControllerInstance = controller.NewRunner(GetErrorHandler(), application.GetManageRunnerService())
		})
	}
	return runnerControllerInstance
}
