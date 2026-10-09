package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var uiStateServiceOnce sync.Once
var uiStateServiceInstance *binding.UIStateService

func GetUIStateService() *binding.UIStateService {
	if uiStateServiceInstance == nil {
		uiStateServiceOnce.Do(func() {
			uiStateServiceInstance = binding.NewUIStateService(controller.GetUIStateController())
		})
	}
	return uiStateServiceInstance
}
