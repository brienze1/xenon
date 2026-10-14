package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var mcpSnippetControllerOnce sync.Once
var mcpSnippetControllerInstance adapter.MCPToolController

func GetMCPSnippetController() adapter.MCPToolController {
	if mcpSnippetControllerInstance == nil {
		mcpSnippetControllerOnce.Do(func() {
			mcpSnippetControllerInstance = controller.NewMCPSnippet(
				GetErrorHandler(),
				application.GetGenerateSnippetService(),
				application.GetSearchTreeService(),
				application.GetSaveRequestService(),
			)
		})
	}
	return mcpSnippetControllerInstance
}
