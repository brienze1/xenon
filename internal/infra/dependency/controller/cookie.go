package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var cookieControllerOnce sync.Once
var cookieControllerInstance adapter.CookieController

func GetCookieController() adapter.CookieController {
	if cookieControllerInstance == nil {
		cookieControllerOnce.Do(func() {
			cookieControllerInstance = controller.NewCookie(GetErrorHandler(), application.GetManageCookiesService())
		})
	}
	return cookieControllerInstance
}
