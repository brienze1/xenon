package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var exampleServiceOnce sync.Once
var exampleServiceInstance *binding.ExampleService

func GetExampleService() *binding.ExampleService {
	if exampleServiceInstance == nil {
		exampleServiceOnce.Do(func() {
			exampleServiceInstance = binding.NewExampleService(controller.GetExampleManageController())
		})
	}
	return exampleServiceInstance
}
