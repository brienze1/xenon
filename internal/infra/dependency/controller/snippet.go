package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var snippetControllerOnce sync.Once
var snippetControllerInstance adapter.SnippetController

func GetSnippetController() adapter.SnippetController {
	if snippetControllerInstance == nil {
		snippetControllerOnce.Do(func() {
			snippetControllerInstance = controller.NewSnippet(GetErrorHandler(), application.GetGenerateSnippetService())
		})
	}
	return snippetControllerInstance
}
