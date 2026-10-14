package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var aiAccessWebserviceOnce sync.Once
var aiAccessWebserviceInstance adapter.AIAccessWebservice

func GetAIAccessWebservice() adapter.AIAccessWebservice {
	if aiAccessWebserviceInstance == nil {
		aiAccessWebserviceOnce.Do(func() {
			aiAccessWebserviceInstance = webservice.NewAIAccessWebservice(external.GetAIAccessListener())
		})
	}
	return aiAccessWebserviceInstance
}
