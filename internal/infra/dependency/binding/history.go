package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var historyBindingOnce sync.Once
var historyBindingInstance *binding.HistoryService

func GetHistoryBinding() *binding.HistoryService {
	if historyBindingInstance == nil {
		historyBindingOnce.Do(func() {
			historyBindingInstance = binding.NewHistoryService(controller.GetHistoryController())
		})
	}
	return historyBindingInstance
}
