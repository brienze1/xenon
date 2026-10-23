package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var startRunOnce sync.Once
var startRunInstance adapter.StartRunService

func GetStartRunService() adapter.StartRunService {
	if startRunInstance == nil {
		startRunOnce.Do(func() {
			startRunInstance = service.NewStartRunService(
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
				integration.GetLocalVariablesPersistence(),
				integration.GetTreePersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetRunPersistence(),
				integration.GetRunPersistence(),
				integration.GetRunEventPublisher(),
				integration.GetIterationDataPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetRunPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
				integration.GetAIApprovalPersistence(),
				integration.GetAIApprovalPersistence(),
				integration.GetAIApprovalTimerWebservice(),
				integration.GetWindowFocusWebservice(),
				integration.GetAIEventPublisher(),
			)
		})
	}
	return startRunInstance
}
