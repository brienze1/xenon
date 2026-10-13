package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var oauthControllerOnce sync.Once
var oauthControllerInstance adapter.OAuthController

func GetOAuthController() adapter.OAuthController {
	if oauthControllerInstance == nil {
		oauthControllerOnce.Do(func() {
			oauthControllerInstance = controller.NewOAuth(GetErrorHandler(), application.GetManageOAuthService())
		})
	}
	return oauthControllerInstance
}
