package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var deleteRunsOnce sync.Once
var deleteRunsInstance adapter.DeleteRunsService

func GetDeleteRunsService() adapter.DeleteRunsService {
	if deleteRunsInstance == nil {
		deleteRunsOnce.Do(func() {
			deleteRunsInstance = service.NewDeleteRunsService(
				integration.GetRegistryPersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetRunPersistence(),
				integration.GetResponseBodyPersistence(),
			)
		})
	}
	return deleteRunsInstance
}
