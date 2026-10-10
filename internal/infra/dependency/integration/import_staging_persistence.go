package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var importStagingPersistenceOnce sync.Once
var importStagingPersistenceInstance adapter.ImportStagingPersistence

func GetImportStagingPersistence() adapter.ImportStagingPersistence {
	if importStagingPersistenceInstance == nil {
		importStagingPersistenceOnce.Do(func() {
			importStagingPersistenceInstance = persistence.NewImportStagingPersistence(GetWorkspaceIndexCache())
		})
	}
	return importStagingPersistenceInstance
}
