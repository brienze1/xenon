package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestSaveControllerOnce sync.Once
var requestSaveControllerInstance adapter.RequestSaveController

func GetRequestSaveController() adapter.RequestSaveController {
	if requestSaveControllerInstance == nil {
		requestSaveControllerOnce.Do(func() {
			requestSaveControllerInstance = controller.NewRequestSave(GetErrorHandler(), application.GetSaveRequestService())
		})
	}
	return requestSaveControllerInstance
}
