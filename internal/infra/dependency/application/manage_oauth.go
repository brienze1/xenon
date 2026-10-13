package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageOAuthOnce sync.Once
var manageOAuthInstance adapter.ManageOAuthService

func GetManageOAuthService() adapter.ManageOAuthService {
	if manageOAuthInstance == nil {
		manageOAuthOnce.Do(func() {
			manageOAuthInstance = service.NewManageOAuthService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetRequestPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKeychainWebservice(),
				integration.GetVariableEventPublisher(),
				integration.GetOAuthBrowserWebservice(),
				integration.GetOAuth2Webservice(),
				integration.GetOAuth2Webservice(),
				integration.GetVaultPersistence(),
				integration.GetClipboardWebservice(),
			)
		})
	}
	return manageOAuthInstance
}
