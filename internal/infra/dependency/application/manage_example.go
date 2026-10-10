package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageExampleOnce sync.Once
var manageExampleInstance adapter.ManageExampleService

func GetManageExampleService() adapter.ManageExampleService {
	if manageExampleInstance == nil {
		manageExampleOnce.Do(func() {
			manageExampleInstance = service.NewManageExampleService(
				integration.GetRegistryPersistence(),
				integration.GetRequestPersistence(),
				integration.GetRequestPersistence(),
				integration.GetResponseBodyPersistence(),
				integration.GetResponseBodyPersistence(),
			)
		})
	}
	return manageExampleInstance
}
