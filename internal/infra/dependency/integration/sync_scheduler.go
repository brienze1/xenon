package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var syncSchedulerOnce sync.Once
var syncSchedulerInstance adapter.SyncScheduler

func GetSyncScheduler() adapter.SyncScheduler {
	if syncSchedulerInstance == nil {
		syncSchedulerOnce.Do(func() {
			syncSchedulerInstance = webservice.NewSyncScheduler(external.GetTimers())
		})
	}
	return syncSchedulerInstance
}
