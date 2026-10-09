package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestResponseOnce sync.Once
var requestResponseInstance adapter.RequestResponseController

func GetRequestResponse() adapter.RequestResponseController {
	if requestResponseInstance == nil {
		requestResponseOnce.Do(func() {
			requestResponseInstance = controller.NewRequestResponse(GetErrorHandler(), application.GetManageResponseService())
		})
	}
	return requestResponseInstance
}
