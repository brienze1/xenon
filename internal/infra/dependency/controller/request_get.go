package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestGetControllerOnce sync.Once
var requestGetControllerInstance adapter.RequestGetController

func GetRequestGetController() adapter.RequestGetController {
	if requestGetControllerInstance == nil {
		requestGetControllerOnce.Do(func() {
			requestGetControllerInstance = controller.NewRequestGet(GetErrorHandler(), application.GetSaveRequestService())
		})
	}
	return requestGetControllerInstance
}
