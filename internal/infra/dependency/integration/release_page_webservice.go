package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var releasePageWebserviceOnce sync.Once
var releasePageWebserviceInstance adapter.ReleasePageWebservice

func GetReleasePageWebservice() adapter.ReleasePageWebservice {
	if releasePageWebserviceInstance == nil {
		releasePageWebserviceOnce.Do(func() {
			releasePageWebserviceInstance = webservice.NewReleasePageWebservice(external.GetBrowser())
		})
	}
	return releasePageWebserviceInstance
}
