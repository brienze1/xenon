package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var historyOnce sync.Once
var historyInstance adapter.HistoryService

func GetHistoryService() adapter.HistoryService {
	if historyInstance == nil {
		historyOnce.Do(func() {
			historyInstance = service.NewHistoryService(
				integration.GetRegistryPersistence(),
				integration.GetHistoryPersistence(),
				integration.GetHistoryPersistence(),
			)
		})
	}
	return historyInstance
}
