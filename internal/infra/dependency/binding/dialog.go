package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var dialogServiceOnce sync.Once
var dialogServiceInstance *binding.DialogService

func GetDialogService() *binding.DialogService {
	if dialogServiceInstance == nil {
		dialogServiceOnce.Do(func() {
			dialogServiceInstance = binding.NewDialogService(controller.GetDialogController())
		})
	}
	return dialogServiceInstance
}
