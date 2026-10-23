package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var planRunOnce sync.Once
var planRunInstance adapter.PlanRunService

func GetPlanRunService() adapter.PlanRunService {
	if planRunInstance == nil {
		planRunOnce.Do(func() {
			planRunInstance = service.NewPlanRunService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetRequestPersistence(),
				integration.GetLocalVariablesPersistence(),
				integration.GetEnvironmentPersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetRunPersistence(),
				integration.GetResponseBodyPersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
				integration.GetUIStatePersistence(),
				integration.GetIterationDataPersistence(),
			)
		})
	}
	return planRunInstance
}
