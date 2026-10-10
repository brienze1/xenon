package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/bruno"
)

var brunoWebserviceOnce sync.Once
var brunoWebserviceInstance adapter.BrunoWebservice

func GetBrunoWebservice() adapter.BrunoWebservice {
	if brunoWebserviceInstance == nil {
		brunoWebserviceOnce.Do(func() {
			brunoWebserviceInstance = bruno.NewBrunoWebservice()
		})
	}
	return brunoWebserviceInstance
}
