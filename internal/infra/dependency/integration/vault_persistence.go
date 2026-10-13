package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var vaultPersistenceOnce sync.Once
var vaultPersistenceInstance adapter.VaultPersistence

func GetVaultPersistence() adapter.VaultPersistence {
	if vaultPersistenceInstance == nil {
		vaultPersistenceOnce.Do(func() {
			vaultPersistenceInstance = persistence.NewVaultPersistence(external.GetKeychain())
		})
	}
	return vaultPersistenceInstance
}
