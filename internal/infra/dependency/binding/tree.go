package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var treeBindingOnce sync.Once
var treeBindingInstance *binding.TreeService

func GetTreeBinding() *binding.TreeService {
	if treeBindingInstance == nil {
		treeBindingOnce.Do(func() {
			treeBindingInstance = binding.NewTreeService(
				controller.GetTreeSearchController(),
				controller.GetTreeManageController(),
				controller.GetTreeSettingsController(),
			)
		})
	}
	return treeBindingInstance
}
