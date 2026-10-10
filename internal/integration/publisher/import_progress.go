package publisher

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type importProgressPublisher struct {
	eventEmitter adapter.EventEmitter
}

func NewImportProgress(eventEmitter adapter.EventEmitter) adapter.ImportProgressPublisher {
	return &importProgressPublisher{eventEmitter: eventEmitter}
}

func (i *importProgressPublisher) Progress(ctx context.Context, progress entity.ImportProgress) error {
	logger.Debug(ctx, "Started", map[string]any{"sourceId": progress.SourceID, "done": progress.Done, "total": progress.Total})

	i.eventEmitter.Emit(dto.ImportProgressEvent, dto.NewImportProgress(progress))

	logger.Debug(ctx, "Finished", map[string]any{"sourceId": progress.SourceID, "done": progress.Done, "total": progress.Total})
	return nil
}
