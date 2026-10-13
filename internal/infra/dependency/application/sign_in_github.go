package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var signInGitHubServiceOnce sync.Once
var signInGitHubServiceInstance adapter.SignInGitHubService

func GetSignInGitHubService() adapter.SignInGitHubService {
	if signInGitHubServiceInstance == nil {
		signInGitHubServiceOnce.Do(func() {
			signInGitHubServiceInstance = service.NewSignInGitHubService(
				integration.GetGitHubWebservice(),
				integration.GetGitHubWebservice(),
				integration.GetGitHubWebservice(),
				integration.GetGitAccountPersistence(),
				integration.GetGitAccountPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKeychainWebservice(),
				integration.GetVariableEventPublisher(),
				integration.GetVaultPersistence(),
			)
		})
	}
	return signInGitHubServiceInstance
}
