package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var gitAccountAddControllerOnce sync.Once
var gitAccountAddControllerInstance adapter.GitAccountAddController

func GetGitAccountAddController() adapter.GitAccountAddController {
	if gitAccountAddControllerInstance == nil {
		gitAccountAddControllerOnce.Do(func() {
			gitAccountAddControllerInstance = controller.NewGitAccountAdd(GetErrorHandler(), application.GetAddGitAccountService())
		})
	}
	return gitAccountAddControllerInstance
}
