package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/persistence"
)

var workspaceIndexOnce sync.Once
var workspaceIndexInstance *persistence.WorkspaceIndex

func GetWorkspaceIndex() *persistence.WorkspaceIndex {
	if workspaceIndexInstance == nil {
		workspaceIndexOnce.Do(func() {
			workspaceIndexInstance = persistence.NewWorkspaceIndex()
		})
	}
	return workspaceIndexInstance
}
