package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var notifyAIChangeOnce sync.Once
var notifyAIChangeInstance adapter.NotifyAIChangeService

func GetNotifyAIChangeService() adapter.NotifyAIChangeService {
	if notifyAIChangeInstance == nil {
		notifyAIChangeOnce.Do(func() {
			notifyAIChangeInstance = service.NewNotifyAIChangeService(integration.GetRegistryPersistence(), integration.GetWorkspaceEventPublisher(), integration.GetVariableEventPublisher(), integration.GetAIEventPublisher())
		})
	}
	return notifyAIChangeInstance
}
