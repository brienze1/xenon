package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpDocsEditControllerOnce sync.Once
var mcpDocsEditControllerInstance adapter.MCPToolController

func GetMCPDocsEditController() adapter.MCPToolController {
	if mcpDocsEditControllerInstance == nil {
		mcpDocsEditControllerOnce.Do(func() {
			mcpDocsEditControllerInstance = controller.NewMCPDocsEdit(
				GetErrorHandler(),
				application.GetSearchTreeService(),
				application.GetManageDocumentationService(),
				application.GetAutoSyncService(),
			)
		})
	}
	return mcpDocsEditControllerInstance
}
