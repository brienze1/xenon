package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var clipboardWebserviceOnce sync.Once
var clipboardWebserviceInstance adapter.ClipboardWebservice

func GetClipboardWebservice() adapter.ClipboardWebservice {
	if clipboardWebserviceInstance == nil {
		clipboardWebserviceOnce.Do(func() {
			clipboardWebserviceInstance = webservice.NewClipboardWebservice(external.GetClipboard())
		})
	}
	return clipboardWebserviceInstance
}
