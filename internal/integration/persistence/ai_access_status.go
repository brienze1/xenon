package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type aiAccessStatusPersistence struct {
	published entity.AIAccessStatus
	mu        sync.Mutex
}

func NewAIAccessStatusPersistence() adapter.AIAccessStatusPersistence {
	return &aiAccessStatusPersistence{published: entity.OffAIAccessStatus(entity.DefaultAIPort)}
}

func (a *aiAccessStatusPersistence) ReplacePublished(ctx context.Context, status entity.AIAccessStatus) entity.AIAccessStatus {
	logger.Debug(ctx, "Started", status)

	a.mu.Lock()
	previous := a.published
	a.published = status
	a.mu.Unlock()

	logger.Debug(ctx, "Finished", previous)
	return previous
}
