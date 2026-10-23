package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var activeRunsPersistenceOnce sync.Once
var activeRunsPersistenceInstance adapter.ActiveRunsPersistence

func GetActiveRunsPersistence() adapter.ActiveRunsPersistence {
	if activeRunsPersistenceInstance == nil {
		activeRunsPersistenceOnce.Do(func() {
			activeRunsPersistenceInstance = persistence.NewActiveRunsPersistence()
		})
	}
	return activeRunsPersistenceInstance
}
