package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var fileChangeOnce sync.Once
var fileChangeInstance adapter.FileChangeController

func GetFileChange() adapter.FileChangeController {
	if fileChangeInstance == nil {
		fileChangeOnce.Do(func() {
			fileChangeInstance = controller.NewFileChange(GetErrorHandler(), application.GetHandleFileChangeService())
		})
	}
	return fileChangeInstance
}
