package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var gitAccountPersistenceOnce sync.Once
var gitAccountPersistenceInstance adapter.GitAccountPersistence

func GetGitAccountPersistence() adapter.GitAccountPersistence {
	if gitAccountPersistenceInstance == nil {
		gitAccountPersistenceOnce.Do(func() {
			gitAccountPersistenceInstance = persistence.NewGitAccountPersistence()
		})
	}
	return gitAccountPersistenceInstance
}
