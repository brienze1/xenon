package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var windowBindingOnce sync.Once
var windowBindingInstance *binding.WindowService

func GetWindowBinding() *binding.WindowService {
	if windowBindingInstance == nil {
		windowBindingOnce.Do(func() {
			windowBindingInstance = binding.NewWindowService(controller.GetWindowStateController())
		})
	}
	return windowBindingInstance
}
