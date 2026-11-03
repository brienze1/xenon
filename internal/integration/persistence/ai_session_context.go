package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type aiSessionContextPersistence struct {
	contexts map[string]entity.AISessionContext
	mu       sync.Mutex
}

func NewAISessionContextPersistence() adapter.AISessionContextPersistence {
	return &aiSessionContextPersistence{contexts: map[string]entity.AISessionContext{}}
}

func (a *aiSessionContextPersistence) Find(ctx context.Context, sessionId string) (entity.AISessionContext, bool) {
	logger.Debug(ctx, "Started", map[string]any{"sessionId": sessionId})

	a.mu.Lock()
	sessionContext, found := a.contexts[sessionId]
	a.mu.Unlock()

	logger.Debug(ctx, "Finished", sessionContext, map[string]any{"sessionId": sessionId, "found": found})
	return sessionContext, found
}

func (a *aiSessionContextPersistence) Save(ctx context.Context, sessionContext entity.AISessionContext) {
	logger.Debug(ctx, "Started", sessionContext)

	if sessionContext.SessionID != "" {
		a.mu.Lock()
		a.contexts[sessionContext.SessionID] = sessionContext
		a.mu.Unlock()
	}

	logger.Debug(ctx, "Finished", sessionContext)
}

func (a *aiSessionContextPersistence) Delete(ctx context.Context, sessionId string) {
	logger.Debug(ctx, "Started", map[string]any{"sessionId": sessionId})

	a.mu.Lock()
	delete(a.contexts, sessionId)
	a.mu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"sessionId": sessionId})
}
