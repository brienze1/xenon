package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var commitChangesServiceOnce sync.Once
var commitChangesServiceInstance adapter.CommitChangesService

func GetCommitChangesService() adapter.CommitChangesService {
	if commitChangesServiceInstance == nil {
		commitChangesServiceOnce.Do(func() {
			commitChangesServiceInstance = service.NewCommitChangesService(
				integration.GetRegistryPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetGitAccountPersistence(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetTreePersistence(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetWorkspacePersistence(),
				integration.GetGitEventPublisher(),
			)
		})
	}
	return commitChangesServiceInstance
}
