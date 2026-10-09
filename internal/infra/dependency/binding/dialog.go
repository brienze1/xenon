package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var dialogBindingOnce sync.Once
var dialogBindingInstance *binding.DialogService

func GetDialogBinding() *binding.DialogService {
	if dialogBindingInstance == nil {
		dialogBindingOnce.Do(func() {
			dialogBindingInstance = binding.NewDialogService(controller.GetDialogController())
		})
	}
	return dialogBindingInstance
}
