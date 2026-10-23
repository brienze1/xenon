package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var controlRunOnce sync.Once
var controlRunInstance adapter.ControlRunService

func GetControlRunService() adapter.ControlRunService {
	if controlRunInstance == nil {
		controlRunOnce.Do(func() {
			controlRunInstance = service.NewControlRunService(
				integration.GetActiveRunsPersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetRunEventPublisher(),
				integration.GetRunPersistence(),
				integration.GetResponseBodyPersistence(),
			)
		})
	}
	return controlRunInstance
}
