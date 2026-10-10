package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var variableTableControllerOnce sync.Once
var variableTableControllerInstance adapter.VariableTableController

func GetVariableTableController() adapter.VariableTableController {
	if variableTableControllerInstance == nil {
		variableTableControllerOnce.Do(func() {
			variableTableControllerInstance = controller.NewVariableTable(GetErrorHandler(), application.GetManageVariableTableService())
		})
	}
	return variableTableControllerInstance
}
