package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/sandbox"
)

var scriptRunnerWebserviceOnce sync.Once
var scriptRunnerWebserviceInstance adapter.ScriptRunnerWebservice

func GetScriptRunnerWebservice() adapter.ScriptRunnerWebservice {
	if scriptRunnerWebserviceInstance == nil {
		scriptRunnerWebserviceOnce.Do(func() {
			scriptRunnerWebserviceInstance = sandbox.NewScriptRunnerWebservice()
		})
	}
	return scriptRunnerWebserviceInstance
}
