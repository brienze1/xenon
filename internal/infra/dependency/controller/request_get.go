package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestGetOnce sync.Once
var requestGetInstance adapter.RequestGetController

func GetRequestGet() adapter.RequestGetController {
	if requestGetInstance == nil {
		requestGetOnce.Do(func() {
			requestGetInstance = controller.NewRequestGet(GetErrorHandler(), application.GetSaveRequestService())
		})
	}
	return requestGetInstance
}
