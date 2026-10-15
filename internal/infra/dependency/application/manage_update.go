package application

import (
	"sync"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/service"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
)

var manageUpdateOnce sync.Once
var manageUpdateInstance adapter.ManageUpdateService

func GetManageUpdateService() adapter.ManageUpdateService {
	if manageUpdateInstance == nil {
		manageUpdateOnce.Do(func() {
			manageUpdateInstance = service.NewManageUpdateService(
				integration.GetUpdateSessionPersistence(),
				integration.GetUpdateSessionPersistence(),
				integration.GetUpdateStatusPublisher(),
				integration.GetUpdateWebservice(),
				integration.GetUpdateWebservice(),
				integration.GetUpdateWebservice(),
				integration.GetUpdateInstallPersistence(),
				integration.GetUpdateInstallPersistence(),
				integration.GetAppRestartWebservice(),
				integration.GetReleasePageWebservice(),
				integration.GetRegistryPersistence(),
			)
		})
	}
	return manageUpdateInstance
}
