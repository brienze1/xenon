package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestSendControllerOnce sync.Once
var requestSendControllerInstance adapter.RequestSendController

func GetRequestSendController() adapter.RequestSendController {
	if requestSendControllerInstance == nil {
		requestSendControllerOnce.Do(func() {
			requestSendControllerInstance = controller.NewRequestSend(GetErrorHandler(), application.GetSendRequestService())
		})
	}
	return requestSendControllerInstance
}
