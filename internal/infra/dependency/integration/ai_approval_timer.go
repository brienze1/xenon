package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var aiApprovalTimerOnce sync.Once
var aiApprovalTimerInstance adapter.AIApprovalTimer

func GetAIApprovalTimer() adapter.AIApprovalTimer {
	if aiApprovalTimerInstance == nil {
		aiApprovalTimerOnce.Do(func() {
			aiApprovalTimerInstance = webservice.NewAIApprovalTimer(external.GetTimers())
		})
	}
	return aiApprovalTimerInstance
}
