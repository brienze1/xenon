package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type updateSessionPersistence struct {
	sessions map[string]entity.UpdateSession
	mu       sync.Mutex
}

// NewUpdateSessionPersistence keeps one in-memory update session per app-data folder.
func NewUpdateSessionPersistence() adapter.UpdateSessionPersistence {
	return &updateSessionPersistence{sessions: map[string]entity.UpdateSession{}}
}

func (u *updateSessionPersistence) FindUpdateSession(ctx context.Context) entity.UpdateSession {
	logger.Debug(ctx, "Started")

	u.mu.Lock()
	session := u.sessions[properties.Properties().Application.DataDir]
	u.mu.Unlock()

	logger.Debug(ctx, "Finished", session)
	return session
}

func (u *updateSessionPersistence) SaveUpdateSession(ctx context.Context, session entity.UpdateSession) {
	logger.Debug(ctx, "Started", session)

	u.mu.Lock()
	u.sessions[properties.Properties().Application.DataDir] = session
	u.mu.Unlock()

	logger.Debug(ctx, "Finished", session)
}

func (u *updateSessionPersistence) ChangeUpdateSession(ctx context.Context, change func(entity.UpdateSession) (entity.UpdateSession, bool)) (entity.UpdateSession, bool) {
	logger.Debug(ctx, "Started")

	u.mu.Lock()
	defer u.mu.Unlock()
	key := properties.Properties().Application.DataDir
	session, changed := change(u.sessions[key])
	if changed {
		u.sessions[key] = session
	}

	logger.Debug(ctx, "Finished", session, map[string]any{"changed": changed})
	return session, changed
}
