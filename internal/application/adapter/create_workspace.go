package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type CreateWorkspaceService interface {
	Create(ctx context.Context, name, path string) (*entity.RegisteredWorkspace, error)
}
