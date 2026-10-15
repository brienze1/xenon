package publisher

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type updateStatusPublisher struct {
	eventEmitter adapter.EventEmitter
}

func NewUpdateStatus(eventEmitter adapter.EventEmitter) adapter.UpdateStatusPublisher {
	return &updateStatusPublisher{eventEmitter: eventEmitter}
}

func (u *updateStatusPublisher) PublishUpdateStatus(ctx context.Context, status entity.UpdateStatus) {
	logger.Debug(ctx, "Started", status)

	u.eventEmitter.Emit(dto.UpdateStatusEvent, dto.NewUpdateStatus(status))

	logger.Debug(ctx, "Finished", status)
}
