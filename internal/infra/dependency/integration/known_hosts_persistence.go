package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var knownHostsPersistenceOnce sync.Once
var knownHostsPersistenceInstance adapter.KnownHostsPersistence

func GetKnownHostsPersistence() adapter.KnownHostsPersistence {
	if knownHostsPersistenceInstance == nil {
		knownHostsPersistenceOnce.Do(func() {
			knownHostsPersistenceInstance = persistence.NewKnownHostsPersistence()
		})
	}
	return knownHostsPersistenceInstance
}
