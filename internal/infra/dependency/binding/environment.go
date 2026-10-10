package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var environmentServiceOnce sync.Once
var environmentServiceInstance *binding.EnvironmentService

func GetEnvironmentService() *binding.EnvironmentService {
	if environmentServiceInstance == nil {
		environmentServiceOnce.Do(func() {
			environmentServiceInstance = binding.NewEnvironmentService(controller.GetEnvironmentListController(), controller.GetEnvironmentSetActiveController())
		})
	}
	return environmentServiceInstance
}
