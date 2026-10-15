package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var aliasServiceOnce sync.Once
var aliasServiceInstance *binding.AliasService

func GetAliasService() *binding.AliasService {
	if aliasServiceInstance == nil {
		aliasServiceOnce.Do(func() {
			aliasServiceInstance = binding.NewAliasService(controller.GetAliasManageController())
		})
	}
	return aliasServiceInstance
}
