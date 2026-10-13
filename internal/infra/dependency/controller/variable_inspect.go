package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var variableInspectControllerOnce sync.Once
var variableInspectControllerInstance adapter.VariableInspectController

func GetVariableInspectController() adapter.VariableInspectController {
	if variableInspectControllerInstance == nil {
		variableInspectControllerOnce.Do(func() {
			variableInspectControllerInstance = controller.NewVariableInspect(GetErrorHandler(), application.GetInspectVariablesService(), application.GetSuggestVariablesService())
		})
	}
	return variableInspectControllerInstance
}
