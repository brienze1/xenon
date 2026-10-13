package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestEditControllerOnce sync.Once
var requestEditControllerInstance adapter.RequestEditController

func GetRequestEditController() adapter.RequestEditController {
	if requestEditControllerInstance == nil {
		requestEditControllerOnce.Do(func() {
			requestEditControllerInstance = controller.NewRequestEdit(GetErrorHandler(), application.GetEditRequestService())
		})
	}
	return requestEditControllerInstance
}
