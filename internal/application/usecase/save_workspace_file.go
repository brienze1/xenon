package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveWorkspaceFile interface {
	Create(ctx context.Context, root string, workspace entity.Workspace) error
}
