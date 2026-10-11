package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var oauthServiceOnce sync.Once
var oauthServiceInstance *binding.OAuthService

func GetOAuthService() *binding.OAuthService {
	if oauthServiceInstance == nil {
		oauthServiceOnce.Do(func() {
			oauthServiceInstance = binding.NewOAuthService(controller.GetOAuthController())
		})
	}
	return oauthServiceInstance
}
