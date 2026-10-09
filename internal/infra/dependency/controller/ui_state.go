package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var uiStateOnce sync.Once
var uiStateInstance adapter.UIStateController

func GetUIState() adapter.UIStateController {
	if uiStateInstance == nil {
		uiStateOnce.Do(func() {
			uiStateInstance = controller.NewUIState(GetErrorHandler(), application.GetUIStateService())
		})
	}
	return uiStateInstance
}
