package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RenameWorkspaceService interface {
	Rename(ctx context.Context, localId, name string) (*entity.RegisteredWorkspace, error)
}
