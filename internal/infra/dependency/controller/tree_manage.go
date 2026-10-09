package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var treeManageOnce sync.Once
var treeManageInstance adapter.TreeManageController

func GetTreeManage() adapter.TreeManageController {
	if treeManageInstance == nil {
		treeManageOnce.Do(func() {
			treeManageInstance = controller.NewTreeManage(GetErrorHandler(), application.GetManageTreeItemService())
		})
	}
	return treeManageInstance
}
