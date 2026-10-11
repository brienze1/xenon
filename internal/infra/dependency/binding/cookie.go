package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var cookieServiceOnce sync.Once
var cookieServiceInstance *binding.CookieService

func GetCookieService() *binding.CookieService {
	if cookieServiceInstance == nil {
		cookieServiceOnce.Do(func() {
			cookieServiceInstance = binding.NewCookieService(controller.GetCookieController())
		})
	}
	return cookieServiceInstance
}
