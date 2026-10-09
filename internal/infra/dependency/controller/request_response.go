package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestResponseControllerOnce sync.Once
var requestResponseControllerInstance adapter.RequestResponseController

func GetRequestResponseController() adapter.RequestResponseController {
	if requestResponseControllerInstance == nil {
		requestResponseControllerOnce.Do(func() {
			requestResponseControllerInstance = controller.NewRequestResponse(GetErrorHandler(), application.GetManageResponseService())
		})
	}
	return requestResponseControllerInstance
}
