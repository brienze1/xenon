package consumer

import (
	"context"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func RegisterFileChangeConsumer() {
	handler := controller.GetFileChange()
	integration.GetWorkspaceWatcherWebservice().Consume(func(localId, path string) {
		_ = handler.Handle(context.Background(), dto.NewFileChange(localId, path))
	})
}
