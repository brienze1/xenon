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
			manageUpdateInstance = service.NewManageUpdateService(service.UpdateUseCases{
				FindUpdateSession:     integration.GetUpdateSessionPersistence(),
				SaveUpdateSession:     integration.GetUpdateSessionPersistence(),
				PublishUpdateStatus:   integration.GetUpdateStatusPublisher(),
				FindLatestRelease:     integration.GetUpdateWebservice(),
				FindReleaseChecksums:  integration.GetUpdateWebservice(),
				DownloadUpdatePackage: integration.GetUpdateWebservice(),
				InstallUpdate:         integration.GetUpdateInstallPersistence(),
				CleanUpdateLeftovers:  integration.GetUpdateInstallPersistence(),
				RestartApp:            integration.GetAppRestartWebservice(),
				OpenReleasePage:       integration.GetUpdateWebservice(),
				FindAppSettings:       integration.GetRegistryPersistence(),
			})
		})
	}
	return manageUpdateInstance
}
