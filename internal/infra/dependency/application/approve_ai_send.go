package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var approveAISendServiceOnce sync.Once
var approveAISendServiceInstance adapter.ApproveAISendService

func GetApproveAISendService() adapter.ApproveAISendService {
	if approveAISendServiceInstance == nil {
		approveAISendServiceOnce.Do(func() {
			approveAISendServiceInstance = service.NewApproveAISendService(
				integration.GetAIApprovalPersistence(),
				integration.GetAIApprovalPersistence(),
				integration.GetAIApprovalPersistence(),
				integration.GetAIApprovalPersistence(),
				integration.GetAIApprovalTimerWebservice(),
				integration.GetWindowFocusWebservice(),
				integration.GetAIEventPublisher(),
			)
		})
	}
	return approveAISendServiceInstance
}
