package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var openApiWebserviceOnce sync.Once
var openApiWebserviceInstance adapter.OpenApiWebservice

func GetOpenApiWebservice() adapter.OpenApiWebservice {
	if openApiWebserviceInstance == nil {
		openApiWebserviceOnce.Do(func() {
			openApiWebserviceInstance = webservice.NewOpenApiWebservice()
		})
	}
	return openApiWebserviceInstance
}
