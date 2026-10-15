package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var updateInstallPersistenceOnce sync.Once
var updateInstallPersistenceInstance adapter.UpdateInstallPersistence

func GetUpdateInstallPersistence() adapter.UpdateInstallPersistence {
	if updateInstallPersistenceInstance == nil {
		updateInstallPersistenceOnce.Do(func() {
			updateInstallPersistenceInstance = persistence.NewUpdateInstallPersistence()
		})
	}
	return updateInstallPersistenceInstance
}
