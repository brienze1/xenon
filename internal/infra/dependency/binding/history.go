package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var historyServiceOnce sync.Once
var historyServiceInstance *binding.HistoryService

func GetHistoryService() *binding.HistoryService {
	if historyServiceInstance == nil {
		historyServiceOnce.Do(func() {
			historyServiceInstance = binding.NewHistoryService(controller.GetHistoryController())
		})
	}
	return historyServiceInstance
}
