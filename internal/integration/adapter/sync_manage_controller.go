package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type SyncManageController interface {
	Status(ctx context.Context, request dto.SyncWorkspaceReference) (*dto.SyncStatus, error)
	Changes(ctx context.Context, request dto.SyncWorkspaceReference) ([]dto.SyncChangeItem, error)
	SaveSettings(ctx context.Context, request dto.SaveSyncSettings) error
}
