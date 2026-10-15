package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var updateWebserviceOnce sync.Once
var updateWebserviceInstance adapter.UpdateWebservice

func GetUpdateWebservice() adapter.UpdateWebservice {
	if updateWebserviceInstance == nil {
		updateWebserviceOnce.Do(func() {
			updateWebserviceInstance = webservice.NewUpdateWebservice(external.GetHttpClientFactory(), external.GetBrowser())
		})
	}
	return updateWebserviceInstance
}
