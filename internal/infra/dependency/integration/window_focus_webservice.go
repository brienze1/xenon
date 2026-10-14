package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var windowFocusWebserviceOnce sync.Once
var windowFocusWebserviceInstance adapter.WindowFocusWebservice

func GetWindowFocusWebservice() adapter.WindowFocusWebservice {
	if windowFocusWebserviceInstance == nil {
		windowFocusWebserviceOnce.Do(func() {
			windowFocusWebserviceInstance = webservice.NewWindowFocusWebservice(external.GetWindowFocuser())
		})
	}
	return windowFocusWebserviceInstance
}
