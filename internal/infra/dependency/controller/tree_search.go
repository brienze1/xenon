package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var treeSearchControllerOnce sync.Once
var treeSearchControllerInstance adapter.TreeSearchController

func GetTreeSearchController() adapter.TreeSearchController {
	if treeSearchControllerInstance == nil {
		treeSearchControllerOnce.Do(func() {
			treeSearchControllerInstance = controller.NewTreeSearch(GetErrorHandler(), application.GetSearchTreeService())
		})
	}
	return treeSearchControllerInstance
}
