package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var recordAIActivityOnce sync.Once
var recordAIActivityInstance adapter.RecordAIActivityService

func GetRecordAIActivityService() adapter.RecordAIActivityService {
	if recordAIActivityInstance == nil {
		recordAIActivityOnce.Do(func() {
			recordAIActivityInstance = service.NewRecordAIActivityService(integration.GetRegistryPersistence(), integration.GetVaultPersistence(), integration.GetKnownSecretsPersistence(), integration.GetHistoryPersistence(), integration.GetAIEventPublisher())
		})
	}
	return recordAIActivityInstance
}
