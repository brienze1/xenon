package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageTreeItemOnce sync.Once
var manageTreeItemInstance adapter.ManageTreeItemService

func GetManageTreeItemService() adapter.ManageTreeItemService {
	if manageTreeItemInstance == nil {
		manageTreeItemOnce.Do(func() {
			manageTreeItemInstance = service.NewManageTreeItemService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
			)
		})
	}
	return manageTreeItemInstance
}
