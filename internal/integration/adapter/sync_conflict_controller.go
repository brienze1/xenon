package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SyncConflictController interface {
	Session(ctx context.Context, request dto.SyncWorkspaceReference) (*dto.SyncSession, error)
	Resolve(ctx context.Context, request dto.ResolveSyncConflict) (*dto.SyncConflictItem, error)
	FinishPull(ctx context.Context, request dto.SyncSessionReference) (*dto.SyncStatus, error)
	CancelPull(ctx context.Context, request dto.SyncSessionReference) (*dto.SyncStatus, error)
}
