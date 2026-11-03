package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var workspaceLockPersistenceOnce sync.Once
var workspaceLockPersistenceInstance adapter.WorkspaceLockPersistence

func GetWorkspaceLockPersistence() adapter.WorkspaceLockPersistence {
	if workspaceLockPersistenceInstance == nil {
		workspaceLockPersistenceOnce.Do(func() {
			workspaceLockPersistenceInstance = persistence.NewWorkspaceLockPersistence()
		})
	}
	return workspaceLockPersistenceInstance
}
