package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var aiApprovalTimerWebserviceOnce sync.Once
var aiApprovalTimerWebserviceInstance adapter.AIApprovalTimerWebservice

func GetAIApprovalTimerWebservice() adapter.AIApprovalTimerWebservice {
	if aiApprovalTimerWebserviceInstance == nil {
		aiApprovalTimerWebserviceOnce.Do(func() {
			aiApprovalTimerWebserviceInstance = webservice.NewAIApprovalTimerWebservice(external.GetTimers())
		})
	}
	return aiApprovalTimerWebserviceInstance
}
