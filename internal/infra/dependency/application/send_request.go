package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var sendRequestOnce sync.Once
var sendRequestInstance adapter.SendRequestService

func GetSendRequestService() adapter.SendRequestService {
	if sendRequestInstance == nil {
		sendRequestOnce.Do(func() {
			sendRequestInstance = service.NewSendRequestService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetRequestPersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetKeychainWebservice(),
				integration.GetVariableEventPublisher(),
				integration.GetRegistryPersistence(),
				integration.GetScriptRunnerWebservice(),
				integration.GetConsolePersistence(),
				integration.GetConsolePersistence(),
				integration.GetConsoleEntriesPublisher(),
				integration.GetDynamicValuesWebservice(),
				integration.GetHttpWebservice(),
				integration.GetResponseBodyPersistence(),
				integration.GetResponseBodyPersistence(),
				integration.GetHistoryPersistence(),
				integration.GetVaultPersistence(),
				integration.GetOAuth2Webservice(),
				integration.GetAwsV4Webservice(),
				integration.GetDigestWebservice(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
			)
		})
	}
	return sendRequestInstance
}
