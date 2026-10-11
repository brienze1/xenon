package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var addGitAccountServiceOnce sync.Once
var addGitAccountServiceInstance adapter.AddGitAccountService

func GetAddGitAccountService() adapter.AddGitAccountService {
	if addGitAccountServiceInstance == nil {
		addGitAccountServiceOnce.Do(func() {
			addGitAccountServiceInstance = service.NewAddGitAccountService(
				integration.GetGitAccountPersistence(),
				integration.GetGitWebservice(),
				integration.GetGitHubWebservice(),
				integration.GetGitWebservice(),
				integration.GetSshKeyPersistence(),
				integration.GetSshKeyPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKeychainWebservice(),
				integration.GetVariableEventPublisher(),
				integration.GetVaultPersistence(),
			)
		})
	}
	return addGitAccountServiceInstance
}
