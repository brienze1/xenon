package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var brunoWebserviceOnce sync.Once
var brunoWebserviceInstance adapter.BrunoWebservice

func GetBrunoWebservice() adapter.BrunoWebservice {
	if brunoWebserviceInstance == nil {
		brunoWebserviceOnce.Do(func() {
			brunoWebserviceInstance = webservice.NewBrunoWebservice()
		})
	}
	return brunoWebserviceInstance
}
