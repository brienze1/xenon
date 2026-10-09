package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var uiStateBindingOnce sync.Once
var uiStateBindingInstance *binding.UIStateService

func GetUIStateBinding() *binding.UIStateService {
	if uiStateBindingInstance == nil {
		uiStateBindingOnce.Do(func() {
			uiStateBindingInstance = binding.NewUIStateService(controller.GetUIStateController())
		})
	}
	return uiStateBindingInstance
}
