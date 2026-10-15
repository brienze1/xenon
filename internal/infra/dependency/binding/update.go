package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var updateServiceOnce sync.Once
var updateServiceInstance *binding.UpdateService

func GetUpdateService() *binding.UpdateService {
	if updateServiceInstance == nil {
		updateServiceOnce.Do(func() {
			updateServiceInstance = binding.NewUpdateService(controller.GetUpdateController())
		})
	}
	return updateServiceInstance
}
