package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var consoleServiceOnce sync.Once
var consoleServiceInstance *binding.ConsoleService

func GetConsoleService() *binding.ConsoleService {
	if consoleServiceInstance == nil {
		consoleServiceOnce.Do(func() {
			consoleServiceInstance = binding.NewConsoleService(controller.GetConsoleController())
		})
	}
	return consoleServiceInstance
}
