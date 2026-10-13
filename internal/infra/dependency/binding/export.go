package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var exportServiceOnce sync.Once
var exportServiceInstance *binding.ExportService

func GetExportService() *binding.ExportService {
	if exportServiceInstance == nil {
		exportServiceOnce.Do(func() {
			exportServiceInstance = binding.NewExportService(controller.GetExportController())
		})
	}
	return exportServiceInstance
}
