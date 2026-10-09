package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var treeServiceOnce sync.Once
var treeServiceInstance *binding.TreeService

func GetTreeService() *binding.TreeService {
	if treeServiceInstance == nil {
		treeServiceOnce.Do(func() {
			treeServiceInstance = binding.NewTreeService(
				controller.GetTreeSearchController(),
				controller.GetTreeManageController(),
				controller.GetTreeSettingsController(),
			)
		})
	}
	return treeServiceInstance
}
