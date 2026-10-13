package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var generateSnippetOnce sync.Once
var generateSnippetInstance adapter.GenerateSnippetService

func GetGenerateSnippetService() adapter.GenerateSnippetService {
	if generateSnippetInstance == nil {
		generateSnippetOnce.Do(func() {
			generateSnippetInstance = service.NewGenerateSnippetService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetWorkspacePersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetVaultPersistence(),
				integration.GetDynamicValuesWebservice(),
				integration.GetAwsV4Webservice(),
				integration.GetDigestWebservice(),
				integration.GetUIStatePersistence(),
				integration.GetSnippetWebservice(),
				integration.GetClipboardWebservice(),
			)
		})
	}
	return generateSnippetInstance
}
