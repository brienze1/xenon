package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var xenonExportPersistenceOnce sync.Once
var xenonExportPersistenceInstance adapter.XenonExportPersistence

func GetXenonExportPersistence() adapter.XenonExportPersistence {
	if xenonExportPersistenceInstance == nil {
		xenonExportPersistenceOnce.Do(func() {
			xenonExportPersistenceInstance = persistence.NewXenonExportPersistence(GetWorkspaceIndexCache())
		})
	}
	return xenonExportPersistenceInstance
}
