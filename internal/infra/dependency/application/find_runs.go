package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var findRunsOnce sync.Once
var findRunsInstance adapter.FindRunsService

func GetFindRunsService() adapter.FindRunsService {
	if findRunsInstance == nil {
		findRunsOnce.Do(func() {
			findRunsInstance = service.NewFindRunsService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetActiveRunsPersistence(),
				integration.GetRunPersistence(),
				integration.GetResponseBodyPersistence(),
			)
		})
	}
	return findRunsInstance
}
