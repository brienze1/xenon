package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var requestEditOnce sync.Once
var requestEditInstance adapter.RequestEditController

func GetRequestEdit() adapter.RequestEditController {
	if requestEditInstance == nil {
		requestEditOnce.Do(func() {
			requestEditInstance = controller.NewRequestEdit(GetErrorHandler(), application.GetEditRequestService())
		})
	}
	return requestEditInstance
}
