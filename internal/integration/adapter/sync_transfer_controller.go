package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SyncTransferController interface {
	Commit(ctx context.Context, request dto.CommitSyncChanges) (*dto.SyncStatus, error)
	Pull(ctx context.Context, request dto.SyncWorkspaceReference) (*dto.SyncPullResult, error)
	Push(ctx context.Context, request dto.SyncWorkspaceReference) (*dto.SyncStatus, error)
}
