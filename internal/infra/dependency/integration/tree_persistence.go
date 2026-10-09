package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var treePersistenceOnce sync.Once
var treePersistenceInstance adapter.TreePersistence

func GetTreePersistence() adapter.TreePersistence {
	if treePersistenceInstance == nil {
		treePersistenceOnce.Do(func() {
			treePersistenceInstance = persistence.NewTreePersistence(GetWorkspaceIndexCache())
		})
	}
	return treePersistenceInstance
}
