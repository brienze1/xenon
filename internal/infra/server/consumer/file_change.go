package consumer

import (
	"context"

	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type fileChangeConsumer struct {
	handler adapter.FileChangeController
	watcher adapter.WorkspaceWatcherConsumerWebservice
}

func NewFileChangeConsumer(handler adapter.FileChangeController, watcher adapter.WorkspaceWatcherConsumerWebservice) serverAdapter.Consumer {
	return &fileChangeConsumer{handler: handler, watcher: watcher}
}

func (f *fileChangeConsumer) Register(ctx context.Context) {
	f.watcher.Consume(ctx, func(localId, path string) {
		_ = f.handler.Handle(context.Background(), dto.NewFileChange(localId, path))
	})
}
