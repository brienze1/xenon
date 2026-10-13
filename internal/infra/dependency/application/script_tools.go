package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var scriptToolsOnce sync.Once
var scriptToolsInstance adapter.ScriptToolsService

func GetScriptToolsService() adapter.ScriptToolsService {
	if scriptToolsInstance == nil {
		scriptToolsOnce.Do(func() {
			scriptToolsInstance = service.NewScriptToolsService(integration.GetScriptToolsWebservice(), integration.GetScriptToolsWebservice(), integration.GetScriptToolsWebservice())
		})
	}
	return scriptToolsInstance
}
