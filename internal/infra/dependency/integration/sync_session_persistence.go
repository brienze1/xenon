package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var syncSessionPersistenceOnce sync.Once
var syncSessionPersistenceInstance adapter.SyncSessionPersistence

func GetSyncSessionPersistence() adapter.SyncSessionPersistence {
	if syncSessionPersistenceInstance == nil {
		syncSessionPersistenceOnce.Do(func() {
			syncSessionPersistenceInstance = persistence.NewSyncSessionPersistence()
		})
	}
	return syncSessionPersistenceInstance
}
