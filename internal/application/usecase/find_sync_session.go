package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindSyncSession interface {
	// FindWorkspaceSession returns nil when the workspace has no open session.
	FindWorkspaceSession(ctx context.Context, localId string) (*entity.SyncSession, error)
	FindSession(ctx context.Context, sessionId string) (*entity.SyncSession, error)
}
