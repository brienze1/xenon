package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var environmentPersistenceOnce sync.Once
var environmentPersistenceInstance adapter.EnvironmentPersistence

func GetEnvironmentPersistence() adapter.EnvironmentPersistence {
	if environmentPersistenceInstance == nil {
		environmentPersistenceOnce.Do(func() {
			environmentPersistenceInstance = persistence.NewEnvironmentPersistence(GetWorkspaceIndexCache())
		})
	}
	return environmentPersistenceInstance
}
