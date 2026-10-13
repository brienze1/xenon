package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var verifyGitHostServiceOnce sync.Once
var verifyGitHostServiceInstance adapter.VerifyGitHostService

func GetVerifyGitHostService() adapter.VerifyGitHostService {
	if verifyGitHostServiceInstance == nil {
		verifyGitHostServiceOnce.Do(func() {
			verifyGitHostServiceInstance = service.NewVerifyGitHostService(
				integration.GetGitAccountPersistence(),
				integration.GetVaultPersistence(),
				integration.GetGitWebservice(),
				integration.GetGitWebservice(),
				integration.GetKnownHostsPersistence(),
				integration.GetKnownHostsPersistence(),
				integration.GetGitEventPublisher(),
			)
		})
	}
	return verifyGitHostServiceInstance
}
