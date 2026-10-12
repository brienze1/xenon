package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var syncServiceOnce sync.Once
var syncServiceInstance *binding.SyncService

func GetSyncService() *binding.SyncService {
	if syncServiceInstance == nil {
		syncServiceOnce.Do(func() {
			syncServiceInstance = binding.NewSyncService(controller.GetSyncLinkController(), controller.GetSyncManageController(), controller.GetSyncTransferController())
		})
	}
	return syncServiceInstance
}
