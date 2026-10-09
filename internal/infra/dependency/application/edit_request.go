package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var editRequestOnce sync.Once
var editRequestInstance adapter.EditRequestService

func GetEditRequestService() adapter.EditRequestService {
	if editRequestInstance == nil {
		editRequestOnce.Do(func() {
			editRequestInstance = service.NewEditRequestService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
			)
		})
	}
	return editRequestInstance
}
