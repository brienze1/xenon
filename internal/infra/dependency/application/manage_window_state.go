package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageWindowStateOnce sync.Once
var manageWindowStateInstance adapter.ManageWindowStateService

func GetManageWindowStateService() adapter.ManageWindowStateService {
	if manageWindowStateInstance == nil {
		manageWindowStateOnce.Do(func() {
			manageWindowStateInstance = service.NewManageWindowStateService(integration.GetRegistryPersistence(), integration.GetRegistryPersistence())
		})
	}
	return manageWindowStateInstance
}
