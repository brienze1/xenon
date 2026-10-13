package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var oauthBrowserWebserviceOnce sync.Once
var oauthBrowserWebserviceInstance adapter.OAuthBrowserWebservice

func GetOAuthBrowserWebservice() adapter.OAuthBrowserWebservice {
	if oauthBrowserWebserviceInstance == nil {
		oauthBrowserWebserviceOnce.Do(func() {
			oauthBrowserWebserviceInstance = webservice.NewOAuthBrowserWebservice(external.GetBrowser(), external.GetLoopback())
		})
	}
	return oauthBrowserWebserviceInstance
}
