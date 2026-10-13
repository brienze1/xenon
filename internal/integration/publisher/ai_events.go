package publisher

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type aiEventPublisher struct {
	eventEmitter adapter.EventEmitter
}

func NewAIEvent(eventEmitter adapter.EventEmitter) adapter.AIEventPublisher {
	return &aiEventPublisher{eventEmitter: eventEmitter}
}

func (a *aiEventPublisher) AccessStatusChanged(ctx context.Context, status entity.AIAccessStatus) error {
	logger.Debug(ctx, "Started", status)

	a.eventEmitter.Emit(dto.AIAccessStatusChangedEvent, dto.NewAIAccessStatusChanged(status))

	logger.Debug(ctx, "Finished", status)
	return nil
}

func (a *aiEventPublisher) ApprovalRequested(ctx context.Context, approval entity.AIApproval) error {
	logger.Debug(ctx, "Started", approval)

	a.eventEmitter.Emit(dto.AIApprovalRequestedEvent, dto.NewAIApprovalRequested(approval))

	logger.Debug(ctx, "Finished", approval)
	return nil
}

func (a *aiEventPublisher) ApprovalResolved(ctx context.Context, approval entity.AIApproval) error {
	logger.Debug(ctx, "Started", approval)

	a.eventEmitter.Emit(dto.AIApprovalResolvedEvent, dto.NewAIApprovalResolved(approval))

	logger.Debug(ctx, "Finished", approval)
	return nil
}

func (a *aiEventPublisher) HistoryChanged(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	a.eventEmitter.Emit(dto.HistoryChangedEvent, dto.NewHistoryChanged(localId))

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (a *aiEventPublisher) WorkspacesChanged(ctx context.Context, activeLocalId string) error {
	logger.Debug(ctx, "Started", map[string]any{"activeLocalId": activeLocalId})

	a.eventEmitter.Emit(dto.WorkspacesChangedEvent, dto.NewWorkspacesChanged(activeLocalId))

	logger.Debug(ctx, "Finished", map[string]any{"activeLocalId": activeLocalId})
	return nil
}
