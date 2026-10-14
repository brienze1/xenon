package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var findAIApprovalsServiceOnce sync.Once
var findAIApprovalsServiceInstance adapter.FindAIApprovalsService

func GetFindAIApprovalsService() adapter.FindAIApprovalsService {
	if findAIApprovalsServiceInstance == nil {
		findAIApprovalsServiceOnce.Do(func() {
			findAIApprovalsServiceInstance = service.NewFindAIApprovalsService(integration.GetAIApprovalPersistence())
		})
	}
	return findAIApprovalsServiceInstance
}
