package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var listWorkspacesOnce sync.Once
var listWorkspacesInstance adapter.ListWorkspacesService

func GetListWorkspacesService() adapter.ListWorkspacesService {
	if listWorkspacesInstance == nil {
		listWorkspacesOnce.Do(func() {
			listWorkspacesInstance = service.NewListWorkspacesService(integration.GetRegistryPersistence())
		})
	}
	return listWorkspacesInstance
}
