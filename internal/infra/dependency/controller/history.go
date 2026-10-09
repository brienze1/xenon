package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var historyOnce sync.Once
var historyInstance adapter.HistoryController

func GetHistory() adapter.HistoryController {
	if historyInstance == nil {
		historyOnce.Do(func() {
			historyInstance = controller.NewHistory(GetErrorHandler(), application.GetHistoryService())
		})
	}
	return historyInstance
}
