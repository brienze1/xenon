package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type CloneWorkspaceService interface {
	Clone(ctx context.Context, settings entity.SyncSettings, accountId, clonePath, name string) (*entity.RegisteredWorkspace, error)
}
