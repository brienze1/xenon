package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var workspaceIndexOnce sync.Once
var workspaceIndexInstance adapter.WorkspaceIndex

func GetWorkspaceIndex() adapter.WorkspaceIndex {
	if workspaceIndexInstance == nil {
		workspaceIndexOnce.Do(func() {
			workspaceIndexInstance = persistence.NewWorkspaceIndex()
		})
	}
	return workspaceIndexInstance
}
