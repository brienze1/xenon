package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageHistoryOnce sync.Once
var manageHistoryInstance adapter.ManageHistoryService

func GetManageHistoryService() adapter.ManageHistoryService {
	if manageHistoryInstance == nil {
		manageHistoryOnce.Do(func() {
			manageHistoryInstance = service.NewManageHistoryService(
				integration.GetRegistryPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetHistoryPersistence(),
				integration.GetHistoryPersistence(),
				integration.GetAIEventPublisher(),
			)
		})
	}
	return manageHistoryInstance
}
