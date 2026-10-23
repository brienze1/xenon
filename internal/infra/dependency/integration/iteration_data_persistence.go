package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var iterationDataPersistenceOnce sync.Once
var iterationDataPersistenceInstance adapter.IterationDataPersistence

func GetIterationDataPersistence() adapter.IterationDataPersistence {
	if iterationDataPersistenceInstance == nil {
		iterationDataPersistenceOnce.Do(func() {
			iterationDataPersistenceInstance = persistence.NewIterationDataPersistence()
		})
	}
	return iterationDataPersistenceInstance
}
