package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var publishAIAccessStatusOnce sync.Once
var publishAIAccessStatusInstance adapter.PublishAIAccessStatusService

func GetPublishAIAccessStatusService() adapter.PublishAIAccessStatusService {
	if publishAIAccessStatusInstance == nil {
		publishAIAccessStatusOnce.Do(func() {
			publishAIAccessStatusInstance = service.NewPublishAIAccessStatusService(
				integration.GetAIAccessWebservice(),
				integration.GetAIAccessStatusPersistence(),
				integration.GetAIEventPublisher(),
			)
		})
	}
	return publishAIAccessStatusInstance
}
