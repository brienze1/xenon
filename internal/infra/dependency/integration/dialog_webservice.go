package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var dialogWebserviceOnce sync.Once
var dialogWebserviceInstance adapter.DialogWebservice

func GetDialogWebservice() adapter.DialogWebservice {
	if dialogWebserviceInstance == nil {
		dialogWebserviceOnce.Do(func() {
			dialogWebserviceInstance = webservice.NewDialogWebservice(external.GetDialog())
		})
	}
	return dialogWebserviceInstance
}
