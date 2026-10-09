package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var requestBindingOnce sync.Once
var requestBindingInstance *binding.RequestService

func GetRequestBinding() *binding.RequestService {
	if requestBindingInstance == nil {
		requestBindingOnce.Do(func() {
			requestBindingInstance = binding.NewRequestService(
				controller.GetRequestGetController(),
				controller.GetRequestSaveController(),
				controller.GetRequestEditController(),
				controller.GetRequestSendController(),
				controller.GetRequestResponseController(),
			)
		})
	}
	return requestBindingInstance
}
