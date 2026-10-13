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
