package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var runPersistenceOnce sync.Once
var runPersistenceInstance adapter.RunPersistence

func GetRunPersistence() adapter.RunPersistence {
	if runPersistenceInstance == nil {
		runPersistenceOnce.Do(func() {
			runPersistenceInstance = persistence.NewRunPersistence()
		})
	}
	return runPersistenceInstance
}
