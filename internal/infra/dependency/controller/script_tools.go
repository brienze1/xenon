package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var scriptToolsControllerOnce sync.Once
var scriptToolsControllerInstance adapter.ScriptToolsController

func GetScriptToolsController() adapter.ScriptToolsController {
	if scriptToolsControllerInstance == nil {
		scriptToolsControllerOnce.Do(func() {
			scriptToolsControllerInstance = controller.NewScriptTools(GetErrorHandler(), application.GetScriptToolsService())
		})
	}
	return scriptToolsControllerInstance
}
