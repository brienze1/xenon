package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var aiAccessStatusPersistenceOnce sync.Once
var aiAccessStatusPersistenceInstance adapter.AIAccessStatusPersistence

func GetAIAccessStatusPersistence() adapter.AIAccessStatusPersistence {
	if aiAccessStatusPersistenceInstance == nil {
		aiAccessStatusPersistenceOnce.Do(func() {
			aiAccessStatusPersistenceInstance = persistence.NewAIAccessStatusPersistence()
		})
	}
	return aiAccessStatusPersistenceInstance
}
