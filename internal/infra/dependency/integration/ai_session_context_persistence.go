package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var aiSessionContextPersistenceOnce sync.Once
var aiSessionContextPersistenceInstance adapter.AISessionContextPersistence

func GetAISessionContextPersistence() adapter.AISessionContextPersistence {
	if aiSessionContextPersistenceInstance == nil {
		aiSessionContextPersistenceOnce.Do(func() {
			aiSessionContextPersistenceInstance = persistence.NewAISessionContextPersistence()
		})
	}
	return aiSessionContextPersistenceInstance
}
