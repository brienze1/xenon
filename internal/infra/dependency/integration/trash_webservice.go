package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var trashWebserviceOnce sync.Once
var trashWebserviceInstance adapter.TrashWebservice

func GetTrashWebservice() adapter.TrashWebservice {
	if trashWebserviceInstance == nil {
		trashWebserviceOnce.Do(func() {
			trashWebserviceInstance = webservice.NewTrashWebservice(external.GetTrash())
		})
	}
	return trashWebserviceInstance
}
