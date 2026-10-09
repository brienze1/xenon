package router

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/binding"
	"github.com/brienze1/xenon/internal/infra/dependency/consumer"
	"github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/infra/server/router"
)

var routerOnce sync.Once
var routerInstance adapter.Router

func GetRouter() adapter.Router {
	if routerInstance == nil {
		routerOnce.Do(func() {
			consumer.GetFileChangeConsumer()
			routerInstance = router.NewRouter(
				binding.GetWorkspaceBinding(),
				binding.GetTreeBinding(),
				binding.GetDialogBinding(),
				binding.GetRequestBinding(),
				binding.GetUIStateBinding(),
				binding.GetHistoryBinding(),
				binding.GetWindowBinding(),
				GetResponseBodyRoute(),
			)
		})
	}
	return routerInstance
}
