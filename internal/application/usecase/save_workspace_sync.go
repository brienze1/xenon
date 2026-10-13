package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveWorkspaceSync interface {
	SaveSync(ctx context.Context, root string, settings entity.SyncSettings) error
}
