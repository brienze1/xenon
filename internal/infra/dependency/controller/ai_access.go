package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var aiAccessControllerOnce sync.Once
var aiAccessControllerInstance adapter.AIAccessController

func GetAIAccessController() adapter.AIAccessController {
	if aiAccessControllerInstance == nil {
		aiAccessControllerOnce.Do(func() {
			aiAccessControllerInstance = controller.NewAIAccess(
				GetErrorHandler(),
				application.GetFindAIAccessStatusService(),
				application.GetManageAIAccessService(),
				application.GetGuardAIOutputService(),
				application.GetNotifyAIChangeService(),
				application.GetRecordAIActivityService(),
				application.GetApproveAISendService(),
				application.GetFindAIApprovalsService(),
				application.GetPublishAIAccessStatusService(),
			)
		})
	}
	return aiAccessControllerInstance
}
