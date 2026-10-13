package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var oauth2WebserviceOnce sync.Once
var oauth2WebserviceInstance adapter.OAuth2Webservice

func GetOAuth2Webservice() adapter.OAuth2Webservice {
	if oauth2WebserviceInstance == nil {
		oauth2WebserviceOnce.Do(func() {
			oauth2WebserviceInstance = webservice.NewOAuth2Webservice(external.GetHttpClientFactory())
		})
	}
	return oauth2WebserviceInstance
}
