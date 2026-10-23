package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var runnerServiceOnce sync.Once
var runnerServiceInstance *binding.RunnerService

func GetRunnerService() *binding.RunnerService {
	if runnerServiceInstance == nil {
		runnerServiceOnce.Do(func() {
			runnerServiceInstance = binding.NewRunnerService(controller.GetRunnerController())
		})
	}
	return runnerServiceInstance
}
