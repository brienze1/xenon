package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var gitWebserviceOnce sync.Once
var gitWebserviceInstance adapter.GitWebservice

func GetGitWebservice() adapter.GitWebservice {
	if gitWebserviceInstance == nil {
		gitWebserviceOnce.Do(func() {
			gitWebserviceInstance = webservice.NewGitWebservice(external.GetSshAgent())
		})
	}
	return gitWebserviceInstance
}
