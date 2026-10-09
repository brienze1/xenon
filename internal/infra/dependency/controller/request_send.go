package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestSendOnce sync.Once
var requestSendInstance adapter.RequestSendController

func GetRequestSend() adapter.RequestSendController {
	if requestSendInstance == nil {
		requestSendOnce.Do(func() {
			requestSendInstance = controller.NewRequestSend(GetErrorHandler(), application.GetSendRequestService())
		})
	}
	return requestSendInstance
}
