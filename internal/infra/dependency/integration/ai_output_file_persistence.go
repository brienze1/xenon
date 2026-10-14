package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var aiOutputFilePersistenceOnce sync.Once
var aiOutputFilePersistenceInstance adapter.AIOutputFilePersistence

func GetAIOutputFilePersistence() adapter.AIOutputFilePersistence {
	if aiOutputFilePersistenceInstance == nil {
		aiOutputFilePersistenceOnce.Do(func() {
			aiOutputFilePersistenceInstance = persistence.NewAIOutputFilePersistence()
		})
	}
	return aiOutputFilePersistenceInstance
}
