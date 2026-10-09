package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var historyControllerOnce sync.Once
var historyControllerInstance adapter.HistoryController

func GetHistoryController() adapter.HistoryController {
	if historyControllerInstance == nil {
		historyControllerOnce.Do(func() {
			historyControllerInstance = controller.NewHistory(GetErrorHandler(), application.GetManageHistoryService())
		})
	}
	return historyControllerInstance
}
