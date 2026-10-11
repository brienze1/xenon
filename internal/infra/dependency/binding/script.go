package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var scriptServiceOnce sync.Once
var scriptServiceInstance *binding.ScriptService

func GetScriptService() *binding.ScriptService {
	if scriptServiceInstance == nil {
		scriptServiceOnce.Do(func() {
			scriptServiceInstance = binding.NewScriptService(controller.GetScriptToolsController())
		})
	}
	return scriptServiceInstance
}
