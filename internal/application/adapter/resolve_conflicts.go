package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type ResolveConflictsService interface {
	// Session returns nil when the workspace has no pull waiting for conflict resolution.
	Session(ctx context.Context, localId string) (*entity.SyncSession, error)
	Resolve(ctx context.Context, sessionId, itemId string, resolution enums.ConflictResolution, fieldPath string) (*entity.ConflictItem, error)
	Cancel(ctx context.Context, sessionId string) (*entity.SyncStatus, error)
}
