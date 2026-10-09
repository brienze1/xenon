package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var historyPersistenceOnce sync.Once
var historyPersistenceInstance adapter.HistoryPersistence

func GetHistoryPersistence() adapter.HistoryPersistence {
	if historyPersistenceInstance == nil {
		historyPersistenceOnce.Do(func() {
			historyPersistenceInstance = persistence.NewHistoryPersistence()
		})
	}
	return historyPersistenceInstance
}
