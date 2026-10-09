package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var workspaceIndexPersistenceOnce sync.Once
var workspaceIndexPersistenceInstance adapter.WorkspaceIndexPersistence

func GetWorkspaceIndexPersistence() adapter.WorkspaceIndexPersistence {
	if workspaceIndexPersistenceInstance == nil {
		workspaceIndexPersistenceOnce.Do(func() {
			workspaceIndexPersistenceInstance = persistence.NewWorkspaceIndexPersistence()
		})
	}
	return workspaceIndexPersistenceInstance
}
