package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var writeAIOutputFileOnce sync.Once
var writeAIOutputFileInstance adapter.WriteAIOutputFileService

func GetWriteAIOutputFileService() adapter.WriteAIOutputFileService {
	if writeAIOutputFileInstance == nil {
		writeAIOutputFileOnce.Do(func() {
			writeAIOutputFileInstance = service.NewWriteAIOutputFileService(integration.GetAIOutputFilePersistence())
		})
	}
	return writeAIOutputFileInstance
}
