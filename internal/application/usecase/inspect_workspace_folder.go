package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type InspectWorkspaceFolder interface {
	Inspect(ctx context.Context, root string) (*entity.FolderInfo, error)
}
