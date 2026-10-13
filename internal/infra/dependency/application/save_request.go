package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var saveRequestOnce sync.Once
var saveRequestInstance adapter.SaveRequestService

func GetSaveRequestService() adapter.SaveRequestService {
	if saveRequestInstance == nil {
		saveRequestOnce.Do(func() {
			saveRequestInstance = service.NewSaveRequestService(
				integration.GetRegistryPersistence(),
				integration.GetRequestPersistence(),
				integration.GetRequestPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
			)
		})
	}
	return saveRequestInstance
}
