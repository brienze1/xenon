package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var registryPersistenceOnce sync.Once
var registryPersistenceInstance adapter.RegistryPersistence

func GetRegistryPersistence() adapter.RegistryPersistence {
	if registryPersistenceInstance == nil {
		registryPersistenceOnce.Do(func() {
			registryPersistenceInstance = persistence.NewRegistryPersistence()
		})
	}
	return registryPersistenceInstance
}
