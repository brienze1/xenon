package persistence

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceLockKey struct {
	localId string
}

type workspaceLockPersistence struct {
	locks sync.Map
}

func NewWorkspaceLockPersistence() adapter.WorkspaceLockPersistence {
	return &workspaceLockPersistence{}
}

func (w *workspaceLockPersistence) LockWorkspace(ctx context.Context, localId string) (context.Context, func()) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	key := workspaceLockKey{localId: localId}
	if held, ok := ctx.Value(key).(*atomic.Bool); ok && held.Load() {
		return ctx, func() {}
	}
	lock, _ := w.locks.LoadOrStore(localId, &sync.Mutex{})
	mutex := lock.(*sync.Mutex)
	mutex.Lock()
	held := &atomic.Bool{}
	held.Store(true)
	var once sync.Once
	unlock := func() {
		once.Do(func() {
			held.Store(false)
			mutex.Unlock()
		})
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return context.WithValue(ctx, key, held), unlock
}
