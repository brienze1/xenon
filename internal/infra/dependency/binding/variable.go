package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var variableServiceOnce sync.Once
var variableServiceInstance *binding.VariableService

func GetVariableService() *binding.VariableService {
	if variableServiceInstance == nil {
		variableServiceOnce.Do(func() {
			variableServiceInstance = binding.NewVariableService(controller.GetVariableTableController())
		})
	}
	return variableServiceInstance
}
