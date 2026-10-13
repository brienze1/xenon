package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var sshKeyPersistenceOnce sync.Once
var sshKeyPersistenceInstance adapter.SshKeyPersistence

func GetSshKeyPersistence() adapter.SshKeyPersistence {
	if sshKeyPersistenceInstance == nil {
		sshKeyPersistenceOnce.Do(func() {
			sshKeyPersistenceInstance = persistence.NewSshKeyPersistence()
		})
	}
	return sshKeyPersistenceInstance
}
