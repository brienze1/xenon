package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var digestWebserviceOnce sync.Once
var digestWebserviceInstance adapter.DigestWebservice

func GetDigestWebservice() adapter.DigestWebservice {
	if digestWebserviceInstance == nil {
		digestWebserviceOnce.Do(func() {
			digestWebserviceInstance = webservice.NewDigestWebservice()
		})
	}
	return digestWebserviceInstance
}
