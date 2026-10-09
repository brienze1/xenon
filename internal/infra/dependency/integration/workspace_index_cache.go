package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var workspaceIndexCacheOnce sync.Once
var workspaceIndexCacheInstance adapter.WorkspaceIndexCache

func GetWorkspaceIndexCache() adapter.WorkspaceIndexCache {
	if workspaceIndexCacheInstance == nil {
		workspaceIndexCacheOnce.Do(func() {
			workspaceIndexCacheInstance = persistence.NewWorkspaceIndexCache()
		})
	}
	return workspaceIndexCacheInstance
}
