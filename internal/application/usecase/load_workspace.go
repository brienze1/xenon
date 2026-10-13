package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type LoadWorkspace interface {
	Load(ctx context.Context, root string) (*entity.Workspace, error)
}
