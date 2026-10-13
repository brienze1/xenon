package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var responseBodyRouteOnce sync.Once
var responseBodyRouteInstance *binding.ResponseBodyRoute

func GetResponseBodyRoute() *binding.ResponseBodyRoute {
	if responseBodyRouteInstance == nil {
		responseBodyRouteOnce.Do(func() {
			responseBodyRouteInstance = binding.NewResponseBodyRoute(controller.GetRequestResponseController())
		})
	}
	return responseBodyRouteInstance
}
