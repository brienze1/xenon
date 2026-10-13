package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type LocateWorkspaceService interface {
	Locate(ctx context.Context, localId, path string) (*entity.RegisteredWorkspace, error)
}
