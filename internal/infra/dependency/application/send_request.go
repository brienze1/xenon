package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
	"github.com/brienze1/xenon/internal/infra/properties"
)

var sendRequestOnce sync.Once
var sendRequestInstance adapter.SendRequestService

func GetSendRequestService() adapter.SendRequestService {
	if sendRequestInstance == nil {
		sendRequestOnce.Do(func() {
			sendRequestInstance = service.NewSendRequestService(
				integration.GetRegistryPersistence(),
				integration.GetTreePersistence(),
				integration.GetHttpWebservice(),
				integration.GetResponseBodyWebservice(),
				integration.GetHistoryPersistence(),
				properties.Properties().Application.Version,
			)
		})
	}
	return sendRequestInstance
}
