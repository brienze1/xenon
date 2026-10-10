package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var consolePersistenceOnce sync.Once
var consolePersistenceInstance adapter.ConsolePersistence

func GetConsolePersistence() adapter.ConsolePersistence {
	if consolePersistenceInstance == nil {
		consolePersistenceOnce.Do(func() {
			consolePersistenceInstance = persistence.NewConsolePersistence()
		})
	}
	return consolePersistenceInstance
}
