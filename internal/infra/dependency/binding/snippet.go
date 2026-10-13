package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var snippetServiceOnce sync.Once
var snippetServiceInstance *binding.SnippetService

func GetSnippetService() *binding.SnippetService {
	if snippetServiceInstance == nil {
		snippetServiceOnce.Do(func() {
			snippetServiceInstance = binding.NewSnippetService(controller.GetSnippetController())
		})
	}
	return snippetServiceInstance
}
