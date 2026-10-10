package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var importSourcePersistenceOnce sync.Once
var importSourcePersistenceInstance adapter.ImportSourcePersistence

func GetImportSourcePersistence() adapter.ImportSourcePersistence {
	if importSourcePersistenceInstance == nil {
		importSourcePersistenceOnce.Do(func() {
			importSourcePersistenceInstance = persistence.NewImportSourcePersistence(external.GetArchive())
		})
	}
	return importSourcePersistenceInstance
}
