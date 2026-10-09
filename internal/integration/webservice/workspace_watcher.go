package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceWatcherWebservice struct {
	fileWatcher adapter.FileWatcher
	consumer    adapter.WorkspaceWatcherConsumerWebservice
}

func NewWorkspaceWatcherWebservice(fileWatcher adapter.FileWatcher, consumer adapter.WorkspaceWatcherConsumerWebservice) adapter.WorkspaceWatcherWebservice {
	return &workspaceWatcherWebservice{fileWatcher: fileWatcher, consumer: consumer}
}

func (w *workspaceWatcherWebservice) Watch(ctx context.Context, localId, root string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "root": root})

	onChange := func(path string, _ bool) { w.consumer.Dispatch(context.Background(), localId, path) }
	if err := w.fileWatcher.Watch(ctx, root, onChange); err != nil {
		return errs.InternalServerError(err, "WKS-30500", "the workspace folder cannot be watched")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "root": root})
	return nil
}

func (w *workspaceWatcherWebservice) Unwatch(ctx context.Context) error {
	logger.Debug(ctx, "Started")

	if err := w.fileWatcher.Stop(); err != nil {
		return errs.InternalServerError(err, "WKS-31500", "the workspace folder watch cannot be stopped")
	}

	logger.Debug(ctx, "Finished")
	return nil
}
