package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageConsoleOnce sync.Once
var manageConsoleInstance adapter.ManageConsoleService

func GetManageConsoleService() adapter.ManageConsoleService {
	if manageConsoleInstance == nil {
		manageConsoleOnce.Do(func() {
			manageConsoleInstance = service.NewManageConsoleService(
				integration.GetConsolePersistence(),
				integration.GetConsolePersistence(),
				integration.GetConsolePersistence(),
				integration.GetConsolePersistence(),
				integration.GetConsoleEntriesPublisher(),
				integration.GetResponseBodyPersistence(),
				integration.GetResponseBodyPersistence(),
			)
		})
	}
	return manageConsoleInstance
}
