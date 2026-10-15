package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var aliasManageControllerOnce sync.Once
var aliasManageControllerInstance adapter.AliasManageController

func GetAliasManageController() adapter.AliasManageController {
	if aliasManageControllerInstance == nil {
		aliasManageControllerOnce.Do(func() {
			aliasManageControllerInstance = controller.NewAliasManage(GetErrorHandler(), application.GetManageAliasService())
		})
	}
	return aliasManageControllerInstance
}
