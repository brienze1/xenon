package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var treeDocumentationControllerOnce sync.Once
var treeDocumentationControllerInstance adapter.TreeDocumentationController

func GetTreeDocumentationController() adapter.TreeDocumentationController {
	if treeDocumentationControllerInstance == nil {
		treeDocumentationControllerOnce.Do(func() {
			treeDocumentationControllerInstance = controller.NewTreeDocumentation(GetErrorHandler(), application.GetManageDocumentationService(), application.GetAutoSyncService())
		})
	}
	return treeDocumentationControllerInstance
}
