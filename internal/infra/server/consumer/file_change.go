package consumer

import (
	"context"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func RegisterFileChangeConsumer() {
	handler := controller.GetFileChangeController()
	integration.GetWorkspaceWatcherConsumer().Consume(context.Background(), func(localId, path string) {
		_ = handler.Handle(context.Background(), dto.NewFileChange(localId, path))
	})
}
