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
