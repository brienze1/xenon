package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var appProcessOnce sync.Once
var appProcessInstance adapter.AppProcess

func GetAppProcess() adapter.AppProcess {
	if appProcessInstance == nil {
		appProcessOnce.Do(func() {
			appProcessInstance = infraAdapter.NewAppProcess()
		})
	}
	return appProcessInstance
}

func SetAppProcess(appProcess adapter.AppProcess) {
	appProcessOnce.Do(func() {})
	appProcessInstance = appProcess
}
