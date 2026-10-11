package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var exportFilePersistenceOnce sync.Once
var exportFilePersistenceInstance adapter.ExportFilePersistence

func GetExportFilePersistence() adapter.ExportFilePersistence {
	if exportFilePersistenceInstance == nil {
		exportFilePersistenceOnce.Do(func() {
			exportFilePersistenceInstance = persistence.NewExportFilePersistence()
		})
	}
	return exportFilePersistenceInstance
}
