package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/pkg/logger"
)

func (v *vaultPersistence) LockToken(ctx context.Context, localId, owner string) func() {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "owner": owner})

	lock, _ := v.tokenLocks.LoadOrStore(localId+"/"+owner, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "owner": owner})
	return mutex.Unlock
}
