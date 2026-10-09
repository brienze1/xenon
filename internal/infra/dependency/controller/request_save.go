package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestSaveOnce sync.Once
var requestSaveInstance adapter.RequestSaveController

func GetRequestSave() adapter.RequestSaveController {
	if requestSaveInstance == nil {
		requestSaveOnce.Do(func() {
			requestSaveInstance = controller.NewRequestSave(GetErrorHandler(), application.GetSaveRequestService())
		})
	}
	return requestSaveInstance
}
