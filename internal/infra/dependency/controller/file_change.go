package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var fileChangeControllerOnce sync.Once
var fileChangeControllerInstance adapter.FileChangeController

func GetFileChangeController() adapter.FileChangeController {
	if fileChangeControllerInstance == nil {
		fileChangeControllerOnce.Do(func() {
			fileChangeControllerInstance = controller.NewFileChange(GetErrorHandler(), application.GetHandleFileChangeService())
		})
	}
	return fileChangeControllerInstance
}
