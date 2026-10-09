package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var requestServiceOnce sync.Once
var requestServiceInstance *binding.RequestService

func GetRequestService() *binding.RequestService {
	if requestServiceInstance == nil {
		requestServiceOnce.Do(func() {
			requestServiceInstance = binding.NewRequestService(
				controller.GetRequestGetController(),
				controller.GetRequestSaveController(),
				controller.GetRequestEditController(),
				controller.GetRequestSendController(),
				controller.GetRequestResponseController(),
			)
		})
	}
	return requestServiceInstance
}
