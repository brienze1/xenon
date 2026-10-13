package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var localVariablesPersistenceOnce sync.Once
var localVariablesPersistenceInstance adapter.LocalVariablesPersistence

func GetLocalVariablesPersistence() adapter.LocalVariablesPersistence {
	if localVariablesPersistenceInstance == nil {
		localVariablesPersistenceOnce.Do(func() {
			localVariablesPersistenceInstance = persistence.NewLocalVariablesPersistence()
		})
	}
	return localVariablesPersistenceInstance
}
