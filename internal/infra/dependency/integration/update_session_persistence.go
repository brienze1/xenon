package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var updateSessionPersistenceOnce sync.Once
var updateSessionPersistenceInstance adapter.UpdateSessionPersistence

func GetUpdateSessionPersistence() adapter.UpdateSessionPersistence {
	if updateSessionPersistenceInstance == nil {
		updateSessionPersistenceOnce.Do(func() {
			updateSessionPersistenceInstance = persistence.NewUpdateSessionPersistence()
		})
	}
	return updateSessionPersistenceInstance
}
