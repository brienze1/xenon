package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var dialogOnce sync.Once
var dialogInstance adapter.Dialog

func GetDialog() adapter.Dialog {
	if dialogInstance == nil {
		dialogOnce.Do(func() {
			dialogInstance = infraAdapter.NewDialog()
		})
	}
	return dialogInstance
}

func SetDialog(dialog adapter.Dialog) {
	dialogOnce.Do(func() {})
	dialogInstance = dialog
}
