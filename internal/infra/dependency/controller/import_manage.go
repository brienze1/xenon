package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var importManageControllerOnce sync.Once
var importManageControllerInstance adapter.ImportManageController

func GetImportManageController() adapter.ImportManageController {
	if importManageControllerInstance == nil {
		importManageControllerOnce.Do(func() {
			importManageControllerInstance = controller.NewImportManage(GetErrorHandler(), application.GetManageImportService())
		})
	}
	return importManageControllerInstance
}
