package publisher

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type gitEventPublisher struct {
	eventEmitter adapter.EventEmitter
}

func NewGitEvent(eventEmitter adapter.EventEmitter) adapter.GitEventPublisher {
	return &gitEventPublisher{eventEmitter: eventEmitter}
}

func (g *gitEventPublisher) HostKeyUnknown(ctx context.Context, repository string, key entity.HostKey) error {
	logger.Debug(ctx, "Started", key, map[string]any{"repository": repository})

	g.eventEmitter.Emit(dto.GitHostKeyUnknownEvent, dto.NewGitHostKeyUnknown(repository, key))

	logger.Debug(ctx, "Finished", key, map[string]any{"repository": repository})
	return nil
}

func (g *gitEventPublisher) StatusChanged(ctx context.Context, localId string, status entity.SyncStatus) error {
	logger.Debug(ctx, "Started", status, map[string]any{"localId": localId})

	g.eventEmitter.Emit(dto.GitStatusChangedEvent, dto.NewGitStatusChanged(localId, status))

	logger.Debug(ctx, "Finished", status, map[string]any{"localId": localId})
	return nil
}

func (g *gitEventPublisher) Progress(ctx context.Context, localId, phase string, done, total int) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "phase": phase, "done": done, "total": total})

	g.eventEmitter.Emit(dto.GitProgressEvent, dto.NewGitProgress(localId, phase, done, total))

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "phase": phase})
	return nil
}
