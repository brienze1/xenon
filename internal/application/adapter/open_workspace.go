package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type OpenWorkspaceService interface {
	Open(ctx context.Context, path string) (*entity.RegisteredWorkspace, error)
}
