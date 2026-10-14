package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/dependency/server"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var aiAccessServiceOnce sync.Once
var aiAccessServiceInstance *binding.AIAccessService

func GetAIAccessService() *binding.AIAccessService {
	if aiAccessServiceInstance == nil {
		aiAccessServiceOnce.Do(func() {
			aiAccessServiceInstance = binding.NewAIAccessService(controller.GetAIAccessController(), server.GetMCPServer())
		})
	}
	return aiAccessServiceInstance
}
