package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var windowServiceOnce sync.Once
var windowServiceInstance *binding.WindowService

func GetWindowService() *binding.WindowService {
	if windowServiceInstance == nil {
		windowServiceOnce.Do(func() {
			windowServiceInstance = binding.NewWindowService(controller.GetWindowStateController())
		})
	}
	return windowServiceInstance
}
