package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var dialogControllerOnce sync.Once
var dialogControllerInstance adapter.DialogController

func GetDialogController() adapter.DialogController {
	if dialogControllerInstance == nil {
		dialogControllerOnce.Do(func() {
			dialogControllerInstance = controller.NewDialog(GetErrorHandler(), application.GetPickPathService())
		})
	}
	return dialogControllerInstance
}
