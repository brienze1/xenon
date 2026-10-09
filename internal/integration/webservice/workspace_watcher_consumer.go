package webservice

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceWatcherConsumer struct {
	consumer func(localId, path string)
	mu       sync.RWMutex
}

func NewWorkspaceWatcherConsumer() adapter.WorkspaceWatcherConsumer {
	return &workspaceWatcherConsumer{}
}

func (w *workspaceWatcherConsumer) Consume(ctx context.Context, consumer func(localId, path string)) {
	logger.Debug(ctx, "Started")

	w.mu.Lock()
	w.consumer = consumer
	w.mu.Unlock()

	logger.Debug(ctx, "Finished")
}

func (w *workspaceWatcherConsumer) Dispatch(ctx context.Context, localId, path string) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "path": path})

	w.mu.RLock()
	consumer := w.consumer
	w.mu.RUnlock()
	if consumer != nil {
		consumer(localId, path)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "path": path})
}
