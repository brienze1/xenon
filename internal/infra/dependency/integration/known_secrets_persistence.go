package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var knownSecretsPersistenceOnce sync.Once
var knownSecretsPersistenceInstance adapter.KnownSecretsPersistence

func GetKnownSecretsPersistence() adapter.KnownSecretsPersistence {
	if knownSecretsPersistenceInstance == nil {
		knownSecretsPersistenceOnce.Do(func() {
			knownSecretsPersistenceInstance = persistence.NewKnownSecretsPersistence()
		})
	}
	return knownSecretsPersistenceInstance
}
