package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var snippetWebserviceOnce sync.Once
var snippetWebserviceInstance adapter.SnippetWebservice

func GetSnippetWebservice() adapter.SnippetWebservice {
	if snippetWebserviceInstance == nil {
		snippetWebserviceOnce.Do(func() {
			snippetWebserviceInstance = webservice.NewSnippetWebservice()
		})
	}
	return snippetWebserviceInstance
}
