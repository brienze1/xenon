package server

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/dependency/middleware"
	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/infra/server/mcp"
)

var mcpRoutesOnce sync.Once
var mcpRoutesInstance serverAdapter.MCPRoutes

func GetMCPRoutes() serverAdapter.MCPRoutes {
	if mcpRoutesInstance == nil {
		mcpRoutesOnce.Do(func() {
			mcpRoutesInstance = mcp.NewRoutes(mcpControllers(), middleware.GetLoopbackGuardMiddleware(), middleware.GetMCPSessionClosingMiddleware())
		})
	}
	return mcpRoutesInstance
}

func mcpControllers() mcp.Controllers {
	return mcp.Controllers{
		Call:          controller.GetMCPCallController(),
		Workspace:     controller.GetMCPWorkspaceController(),
		WorkspaceEdit: controller.GetMCPWorkspaceEditController(),
		Tree:          controller.GetMCPTreeController(),
		TreeEdit:      controller.GetMCPTreeEditController(),
		Env:           controller.GetMCPEnvController(),
		EnvEdit:       controller.GetMCPEnvEditController(),
		History:       controller.GetMCPHistoryController(),
		Docs:          controller.GetMCPDocsController(),
		DocsEdit:      controller.GetMCPDocsEditController(),
		Response:      controller.GetMCPResponseController(),
		Snippet:       controller.GetMCPSnippetController(),
		Send:          controller.GetMCPSendController(),
		Cookies:       controller.GetMCPCookiesController(),
		CookiesEdit:   controller.GetMCPCookiesEditController(),
		Console:       controller.GetMCPConsoleController(),
		Git:           controller.GetMCPGitController(),
		GitEdit:       controller.GetMCPGitEditController(),
		GitPush:       controller.GetMCPGitPushController(),
		Scripts:       controller.GetMCPScriptsController(),
		Settings:      controller.GetMCPSettingsController(),
		SettingsEdit:  controller.GetMCPSettingsEditController(),
		Import:        controller.GetMCPImportController(),
		Export:        controller.GetMCPExportController(),
		Delete:        controller.GetMCPDeleteController(),
	}
}
