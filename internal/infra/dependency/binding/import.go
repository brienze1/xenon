package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var importServiceOnce sync.Once
var importServiceInstance *binding.ImportService

func GetImportService() *binding.ImportService {
	if importServiceInstance == nil {
		importServiceOnce.Do(func() {
			importServiceInstance = binding.NewImportService(controller.GetImportManageController())
		})
	}
	return importServiceInstance
}
