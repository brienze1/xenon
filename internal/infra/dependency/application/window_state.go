package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var windowStateOnce sync.Once
var windowStateInstance adapter.WindowStateService

func GetWindowStateService() adapter.WindowStateService {
	if windowStateInstance == nil {
		windowStateOnce.Do(func() {
			windowStateInstance = service.NewWindowStateService(integration.GetRegistryPersistence(), integration.GetRegistryPersistence())
		})
	}
	return windowStateInstance
}
