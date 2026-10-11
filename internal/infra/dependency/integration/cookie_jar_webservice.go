package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var cookieJarWebserviceOnce sync.Once
var cookieJarWebserviceInstance adapter.CookieJarWebservice

func GetCookieJarWebservice() adapter.CookieJarWebservice {
	if cookieJarWebserviceInstance == nil {
		cookieJarWebserviceOnce.Do(func() {
			cookieJarWebserviceInstance = webservice.NewCookieJarWebservice()
		})
	}
	return cookieJarWebserviceInstance
}
