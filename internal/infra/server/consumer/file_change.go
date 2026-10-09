package consumer

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type FileChangeConsumer struct {
	handler adapter.FileChangeController
	watcher adapter.WorkspaceWatcherConsumerWebservice
}

func NewFileChangeConsumer(handler adapter.FileChangeController, watcher adapter.WorkspaceWatcherConsumerWebservice) *FileChangeConsumer {
	return &FileChangeConsumer{handler: handler, watcher: watcher}
}

func (f *FileChangeConsumer) Register(ctx context.Context) {
	f.watcher.Consume(ctx, func(localId, path string) {
		_ = f.handler.Handle(context.Background(), dto.NewFileChange(localId, path))
	})
}
