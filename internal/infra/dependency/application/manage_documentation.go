package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageDocumentationOnce sync.Once
var manageDocumentationInstance adapter.ManageDocumentationService

func GetManageDocumentationService() adapter.ManageDocumentationService {
	if manageDocumentationInstance == nil {
		manageDocumentationOnce.Do(func() {
			manageDocumentationInstance = service.NewManageDocumentationService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetRequestPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetRequestPersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetTreePersistence(),
				integration.GetOpenApiWebservice(),
				integration.GetOpenApiWebservice(),
			)
		})
	}
	return manageDocumentationInstance
}
