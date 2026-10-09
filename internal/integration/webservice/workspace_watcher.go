package webservice

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceWatcherWebservice struct {
	fileWatcher adapter.FileWatcher
	consumer    func(localId, path string)
	mu          sync.RWMutex
}

func NewWorkspaceWatcherWebservice(fileWatcher adapter.FileWatcher) adapter.WorkspaceWatcherWebservice {
	return &workspaceWatcherWebservice{fileWatcher: fileWatcher}
}

func (w *workspaceWatcherWebservice) Consume(consumer func(localId, path string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.consumer = consumer
}

func (w *workspaceWatcherWebservice) Watch(ctx context.Context, localId, root string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "root": root})

	if err := w.fileWatcher.Watch(ctx, root, func(path string, _ bool) { w.dispatch(localId, path) }); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "root": root})
	return nil
}

func (w *workspaceWatcherWebservice) Unwatch(ctx context.Context) error {
	logger.Debug(ctx, "Started")

	if err := w.fileWatcher.Stop(); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished")
	return nil
}

func (w *workspaceWatcherWebservice) dispatch(localId, path string) {
	w.mu.RLock()
	consumer := w.consumer
	w.mu.RUnlock()
	if consumer != nil {
		consumer(localId, path)
	}
}
