package persistence

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type syncSessionPersistence struct {
	sessions map[string]entity.SyncSession
	mu       sync.Mutex
}

// NewSyncSessionPersistence keeps open pull sessions in memory only: nothing was written, so quitting discards them safely.
func NewSyncSessionPersistence() adapter.SyncSessionPersistence {
	return &syncSessionPersistence{sessions: map[string]entity.SyncSession{}}
}

func (s *syncSessionPersistence) SaveSession(ctx context.Context, session entity.SyncSession) error {
	logger.Debug(ctx, "Started", session)

	s.mu.Lock()
	s.sessions[session.LocalID] = session
	s.mu.Unlock()

	logger.Debug(ctx, "Finished", session)
	return nil
}

func (s *syncSessionPersistence) UpdateSession(ctx context.Context, session entity.SyncSession) error {
	logger.Debug(ctx, "Started", session)

	s.mu.Lock()
	defer s.mu.Unlock()
	if current, found := s.sessions[session.LocalID]; !found || current.ID != session.ID {
		return errs.NotFoundError("the pull "+session.ID+" is not open any more", "GIT-38404")
	}
	s.sessions[session.LocalID] = session

	logger.Debug(ctx, "Finished", session)
	return nil
}

func (s *syncSessionPersistence) DiscardSession(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	s.mu.Lock()
	delete(s.sessions, localId)
	s.mu.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (s *syncSessionPersistence) FindWorkspaceSession(ctx context.Context, localId string) (*entity.SyncSession, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	s.mu.Lock()
	session, found := s.sessions[localId]
	s.mu.Unlock()
	if !found {
		logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "found": false})
		return nil, nil
	}

	logger.Debug(ctx, "Finished", session)
	return &session, nil
}

func (s *syncSessionPersistence) FindSession(ctx context.Context, sessionId string) (*entity.SyncSession, error) {
	logger.Debug(ctx, "Started", map[string]any{"sessionId": sessionId})

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, session := range s.sessions {
		if session.ID == sessionId {
			logger.Debug(ctx, "Finished", session)
			return &session, nil
		}
	}
	return nil, errs.NotFoundError("the pull "+sessionId+" is not open any more", "GIT-16404")
}
