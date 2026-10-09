package router

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/router"
)

var responseBodyRouteOnce sync.Once
var responseBodyRouteInstance *router.ResponseBodyRoute

func GetResponseBodyRoute() *router.ResponseBodyRoute {
	if responseBodyRouteInstance == nil {
		responseBodyRouteOnce.Do(func() {
			responseBodyRouteInstance = router.NewResponseBodyRoute(controller.GetRequestResponseController())
		})
	}
	return responseBodyRouteInstance
}
