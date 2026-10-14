package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var findAIAccessStatusServiceOnce sync.Once
var findAIAccessStatusServiceInstance adapter.FindAIAccessStatusService

func GetFindAIAccessStatusService() adapter.FindAIAccessStatusService {
	if findAIAccessStatusServiceInstance == nil {
		findAIAccessStatusServiceOnce.Do(func() {
			findAIAccessStatusServiceInstance = service.NewFindAIAccessStatusService(integration.GetAIAccessWebservice())
		})
	}
	return findAIAccessStatusServiceInstance
}
