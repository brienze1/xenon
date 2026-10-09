package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var errorHandlerOnce sync.Once
var errorHandlerInstance adapter.ErrorHandler

func GetErrorHandler() adapter.ErrorHandler {
	if errorHandlerInstance == nil {
		errorHandlerOnce.Do(func() {
			errorHandlerInstance = controller.NewErrorHandler()
		})
	}
	return errorHandlerInstance
}
