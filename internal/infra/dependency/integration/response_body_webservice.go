package integration

import (
	"os"
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var responseBodyWebserviceOnce sync.Once
var responseBodyWebserviceInstance adapter.ResponseBodyWebservice

func GetResponseBodyWebservice() adapter.ResponseBodyWebservice {
	if responseBodyWebserviceInstance == nil {
		responseBodyWebserviceOnce.Do(func() {
			responseBodyWebserviceInstance = webservice.NewResponseBodyWebservice(os.TempDir(), os.Getpid())
		})
	}
	return responseBodyWebserviceInstance
}
