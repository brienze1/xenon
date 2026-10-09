package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var pickPathOnce sync.Once
var pickPathInstance adapter.PickPathService

func GetPickPathService() adapter.PickPathService {
	if pickPathInstance == nil {
		pickPathOnce.Do(func() {
			pickPathInstance = service.NewPickPathService(integration.GetDialogWebservice())
		})
	}
	return pickPathInstance
}
