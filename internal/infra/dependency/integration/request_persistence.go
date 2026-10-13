package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var requestPersistenceOnce sync.Once
var requestPersistenceInstance adapter.RequestPersistence

func GetRequestPersistence() adapter.RequestPersistence {
	if requestPersistenceInstance == nil {
		requestPersistenceOnce.Do(func() {
			requestPersistenceInstance = persistence.NewRequestPersistence(GetWorkspaceIndexCache())
		})
	}
	return requestPersistenceInstance
}
