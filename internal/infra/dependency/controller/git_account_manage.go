package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var gitAccountManageControllerOnce sync.Once
var gitAccountManageControllerInstance adapter.GitAccountManageController

func GetGitAccountManageController() adapter.GitAccountManageController {
	if gitAccountManageControllerInstance == nil {
		gitAccountManageControllerOnce.Do(func() {
			gitAccountManageControllerInstance = controller.NewGitAccountManage(GetErrorHandler(), application.GetManageGitAccountsService())
		})
	}
	return gitAccountManageControllerInstance
}
