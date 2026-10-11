package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageGitAccountsServiceOnce sync.Once
var manageGitAccountsServiceInstance adapter.ManageGitAccountsService

func GetManageGitAccountsService() adapter.ManageGitAccountsService {
	if manageGitAccountsServiceInstance == nil {
		manageGitAccountsServiceOnce.Do(func() {
			manageGitAccountsServiceInstance = service.NewManageGitAccountsService(
				integration.GetGitAccountPersistence(),
				integration.GetGitAccountPersistence(),
				integration.GetGitAccountPersistence(),
				integration.GetVaultPersistence(),
				integration.GetSshKeyPersistence(),
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
			)
		})
	}
	return manageGitAccountsServiceInstance
}
