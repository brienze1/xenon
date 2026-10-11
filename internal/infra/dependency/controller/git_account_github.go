package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var gitAccountGitHubControllerOnce sync.Once
var gitAccountGitHubControllerInstance adapter.GitAccountGitHubController

func GetGitAccountGitHubController() adapter.GitAccountGitHubController {
	if gitAccountGitHubControllerInstance == nil {
		gitAccountGitHubControllerOnce.Do(func() {
			gitAccountGitHubControllerInstance = controller.NewGitAccountGitHub(GetErrorHandler(), application.GetSignInGitHubService())
		})
	}
	return gitAccountGitHubControllerInstance
}
