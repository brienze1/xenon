package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/sandbox"
)

var scriptToolsWebserviceOnce sync.Once
var scriptToolsWebserviceInstance adapter.ScriptToolsWebservice

func GetScriptToolsWebservice() adapter.ScriptToolsWebservice {
	if scriptToolsWebserviceInstance == nil {
		scriptToolsWebserviceOnce.Do(func() {
			scriptToolsWebserviceInstance = sandbox.NewScriptToolsWebservice()
		})
	}
	return scriptToolsWebserviceInstance
}
