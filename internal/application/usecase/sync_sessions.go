package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveSyncSession interface {
	SaveSession(ctx context.Context, session entity.SyncSession) error
	// UpdateSession refuses a session that is no longer the workspace's open one.
	UpdateSession(ctx context.Context, session entity.SyncSession) error
	DiscardSession(ctx context.Context, localId string) error
}

type FindSyncSession interface {
	// FindWorkspaceSession returns nil when the workspace has no open session.
	FindWorkspaceSession(ctx context.Context, localId string) (*entity.SyncSession, error)
	FindSession(ctx context.Context, sessionId string) (*entity.SyncSession, error)
}
