package router

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/binding"
	"github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/infra/server/router"
)

var routerOnce sync.Once
var routerInstance adapter.Router

func GetRouter() adapter.Router {
	if routerInstance == nil {
		routerOnce.Do(func() {
			routerInstance = router.NewRouter(
				binding.GetWorkspaceService(),
				binding.GetTreeService(),
				binding.GetDialogService(),
				binding.GetRequestService(),
				binding.GetUIStateService(),
				binding.GetHistoryService(),
				binding.GetWindowService(),
				binding.GetResponseBodyRoute(),
				binding.GetEnvironmentService(),
				binding.GetVariableService(),
				binding.GetVaultService(),
				binding.GetImportService(),
				binding.GetExampleService(),
				binding.GetSettingsService(),
				binding.GetConsoleService(),
				binding.GetScriptService(),
			)
		})
	}
	return routerInstance
}
