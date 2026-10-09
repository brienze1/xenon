package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var treeManageControllerOnce sync.Once
var treeManageControllerInstance adapter.TreeManageController

func GetTreeManageController() adapter.TreeManageController {
	if treeManageControllerInstance == nil {
		treeManageControllerOnce.Do(func() {
			treeManageControllerInstance = controller.NewTreeManage(GetErrorHandler(), application.GetManageTreeItemService())
		})
	}
	return treeManageControllerInstance
}
