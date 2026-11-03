package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageAISessionOnce sync.Once
var manageAISessionInstance adapter.ManageAISessionService

func GetManageAISessionService() adapter.ManageAISessionService {
	if manageAISessionInstance == nil {
		manageAISessionOnce.Do(func() {
			manageAISessionInstance = service.NewManageAISessionService(
				integration.GetRegistryPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetAISessionContextPersistence(),
				integration.GetAISessionContextPersistence(),
				integration.GetAISessionContextPersistence(),
			)
		})
	}
	return manageAISessionInstance
}
