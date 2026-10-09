package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var workspacePersistenceOnce sync.Once
var workspacePersistenceInstance adapter.WorkspacePersistence

func GetWorkspacePersistence() adapter.WorkspacePersistence {
	if workspacePersistenceInstance == nil {
		workspacePersistenceOnce.Do(func() {
			workspacePersistenceInstance = persistence.NewWorkspacePersistence()
		})
	}
	return workspacePersistenceInstance
}
