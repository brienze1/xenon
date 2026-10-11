package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var gitAccountServiceOnce sync.Once
var gitAccountServiceInstance *binding.GitAccountService

func GetGitAccountService() *binding.GitAccountService {
	if gitAccountServiceInstance == nil {
		gitAccountServiceOnce.Do(func() {
			gitAccountServiceInstance = binding.NewGitAccountService(controller.GetGitAccountManageController(), controller.GetGitAccountAddController(), controller.GetGitAccountGitHubController(), controller.GetGitAccountHostController())
		})
	}
	return gitAccountServiceInstance
}
