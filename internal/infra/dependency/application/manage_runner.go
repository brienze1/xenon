package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageRunnerOnce sync.Once
var manageRunnerInstance adapter.ManageRunnerService

func GetManageRunnerService() adapter.ManageRunnerService {
	if manageRunnerInstance == nil {
		manageRunnerOnce.Do(func() {
			manageRunnerInstance = service.NewManageRunnerService(
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
				integration.GetCookieJarWebservice(),
				integration.GetVaultPersistence(),
				integration.GetVaultPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetWorkspaceEventPublisher(),
				integration.GetKnownSecretsPersistence(),
				integration.GetAIEventPublisher(),
				integration.GetLocalVariablesPersistence(),
				integration.GetTreePersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetRunPersistence(),
				integration.GetRunPersistence(),
				integration.GetRunEventPublisher(),
				integration.GetRunPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
				integration.GetIterationDataPersistence(),
				integration.GetResponseBodyPersistence(),
				integration.GetExportFilePersistence(),
			)
		})
	}
	return manageRunnerInstance
}
