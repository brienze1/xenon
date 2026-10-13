package publisher

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type exportProgressPublisher struct {
	eventEmitter adapter.EventEmitter
}

func NewExportProgress(eventEmitter adapter.EventEmitter) adapter.ExportProgressPublisher {
	return &exportProgressPublisher{eventEmitter: eventEmitter}
}

func (e *exportProgressPublisher) Progress(ctx context.Context, progress entity.ExportProgress) error {
	logger.Debug(ctx, "Started", map[string]any{"done": progress.Done, "total": progress.Total})

	e.eventEmitter.Emit(dto.ExportProgressEvent, dto.NewExportProgress(progress))

	logger.Debug(ctx, "Finished", map[string]any{"done": progress.Done, "total": progress.Total})
	return nil
}
