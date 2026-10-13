package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var gitAccountHostControllerOnce sync.Once
var gitAccountHostControllerInstance adapter.GitAccountHostController

func GetGitAccountHostController() adapter.GitAccountHostController {
	if gitAccountHostControllerInstance == nil {
		gitAccountHostControllerOnce.Do(func() {
			gitAccountHostControllerInstance = controller.NewGitAccountHost(GetErrorHandler(), application.GetVerifyGitHostService())
		})
	}
	return gitAccountHostControllerInstance
}
