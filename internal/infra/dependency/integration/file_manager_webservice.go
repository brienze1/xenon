package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var fileManagerWebserviceOnce sync.Once
var fileManagerWebserviceInstance adapter.FileManagerWebservice

func GetFileManagerWebservice() adapter.FileManagerWebservice {
	if fileManagerWebserviceInstance == nil {
		fileManagerWebserviceOnce.Do(func() {
			fileManagerWebserviceInstance = webservice.NewFileManagerWebservice(external.GetFileManager())
		})
	}
	return fileManagerWebserviceInstance
}
