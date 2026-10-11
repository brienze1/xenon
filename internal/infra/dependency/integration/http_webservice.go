package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var httpWebserviceOnce sync.Once
var httpWebserviceInstance adapter.HttpWebservice

func GetHttpWebservice() adapter.HttpWebservice {
	if httpWebserviceInstance == nil {
		httpWebserviceOnce.Do(func() {
			httpWebserviceInstance = webservice.NewHttpWebservice(external.GetHttpClientFactory(), GetCookieJarWebservice())
		})
	}
	return httpWebserviceInstance
}
