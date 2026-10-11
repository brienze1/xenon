package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var gitHubWebserviceOnce sync.Once
var gitHubWebserviceInstance adapter.GitHubWebservice

func GetGitHubWebservice() adapter.GitHubWebservice {
	if gitHubWebserviceInstance == nil {
		gitHubWebserviceOnce.Do(func() {
			gitHubWebserviceInstance = webservice.NewGitHubWebservice(external.GetHttpClientFactory(), external.GetBrowser())
		})
	}
	return gitHubWebserviceInstance
}
