package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageAIAccessOnce sync.Once
var manageAIAccessInstance adapter.ManageAIAccessService

func GetManageAIAccessService() adapter.ManageAIAccessService {
	if manageAIAccessInstance == nil {
		manageAIAccessOnce.Do(func() {
			manageAIAccessInstance = service.NewManageAIAccessService(
				integration.GetRegistryPersistence(),
				integration.GetAIAccessWebservice(),
				integration.GetAIAccessStatusPersistence(),
				integration.GetAIEventPublisher(),
			)
		})
	}
	return manageAIAccessInstance
}
