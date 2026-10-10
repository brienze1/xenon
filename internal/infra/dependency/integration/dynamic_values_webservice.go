package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var dynamicValuesWebserviceOnce sync.Once
var dynamicValuesWebserviceInstance adapter.DynamicValuesWebservice

func GetDynamicValuesWebservice() adapter.DynamicValuesWebservice {
	if dynamicValuesWebserviceInstance == nil {
		dynamicValuesWebserviceOnce.Do(func() {
			dynamicValuesWebserviceInstance = webservice.NewDynamicValuesWebservice(external.GetRandomSource())
		})
	}
	return dynamicValuesWebserviceInstance
}
