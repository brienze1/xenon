package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var uiStatePersistenceOnce sync.Once
var uiStatePersistenceInstance adapter.UIStatePersistence

func GetUIStatePersistence() adapter.UIStatePersistence {
	if uiStatePersistenceInstance == nil {
		uiStatePersistenceOnce.Do(func() {
			uiStatePersistenceInstance = persistence.NewUIStatePersistence()
		})
	}
	return uiStatePersistenceInstance
}
