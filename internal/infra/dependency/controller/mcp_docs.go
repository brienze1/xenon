package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpDocsControllerOnce sync.Once
var mcpDocsControllerInstance adapter.MCPToolController

func GetMCPDocsController() adapter.MCPToolController {
	if mcpDocsControllerInstance == nil {
		mcpDocsControllerOnce.Do(func() {
			mcpDocsControllerInstance = controller.NewMCPDocs(
				GetErrorHandler(),
				application.GetSearchTreeService(),
				application.GetManageDocumentationService(),
			)
		})
	}
	return mcpDocsControllerInstance
}
