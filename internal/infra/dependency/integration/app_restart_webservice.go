package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var appRestartWebserviceOnce sync.Once
var appRestartWebserviceInstance adapter.AppRestartWebservice

func GetAppRestartWebservice() adapter.AppRestartWebservice {
	if appRestartWebserviceInstance == nil {
		appRestartWebserviceOnce.Do(func() {
			appRestartWebserviceInstance = webservice.NewAppRestartWebservice(external.GetAppProcess())
		})
	}
	return appRestartWebserviceInstance
}
