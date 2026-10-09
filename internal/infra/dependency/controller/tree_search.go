package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var treeSearchOnce sync.Once
var treeSearchInstance adapter.TreeSearchController

func GetTreeSearch() adapter.TreeSearchController {
	if treeSearchInstance == nil {
		treeSearchOnce.Do(func() {
			treeSearchInstance = controller.NewTreeSearch(GetErrorHandler(), application.GetSearchTreeService())
		})
	}
	return treeSearchInstance
}
