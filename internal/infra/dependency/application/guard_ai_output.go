package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var guardAIOutputOnce sync.Once
var guardAIOutputInstance adapter.GuardAIOutputService

func GetGuardAIOutputService() adapter.GuardAIOutputService {
	if guardAIOutputInstance == nil {
		guardAIOutputOnce.Do(func() {
			guardAIOutputInstance = service.NewGuardAIOutputService(integration.GetVaultPersistence(), integration.GetKnownSecretsPersistence())
		})
	}
	return guardAIOutputInstance
}
