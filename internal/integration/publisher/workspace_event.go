package publisher

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type workspaceEventPublisher struct {
	eventEmitter adapter.EventEmitter
}

func NewWorkspaceEvent(eventEmitter adapter.EventEmitter) adapter.WorkspaceEventPublisher {
	return &workspaceEventPublisher{eventEmitter: eventEmitter}
}

func (w *workspaceEventPublisher) TreeChanged(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	w.eventEmitter.Emit(dto.WorkspaceTreeChangedEvent, dto.NewWorkspaceTreeChanged(localId))

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (w *workspaceEventPublisher) RequestChangedOnDisk(ctx context.Context, localId string, change entity.RequestChange) error {
	logger.Debug(ctx, "Started", change, map[string]any{"localId": localId})

	w.eventEmitter.Emit(dto.RequestChangedOnDiskEvent, dto.NewRequestChangedOnDisk(localId, change))

	logger.Debug(ctx, "Finished", change, map[string]any{"localId": localId})
	return nil
}

func (w *workspaceEventPublisher) FolderMissing(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	w.eventEmitter.Emit(dto.WorkspaceFolderMissingEvent, dto.NewWorkspaceFolderMissing(localId))

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (w *workspaceEventPublisher) CookiesChanged(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	w.eventEmitter.Emit(dto.CookiesChangedEvent, dto.NewCookiesChanged(localId))

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}
