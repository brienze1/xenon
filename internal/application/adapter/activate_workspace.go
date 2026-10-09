package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ActivateWorkspaceService interface {
	Activate(ctx context.Context, localId string) (*entity.RegisteredWorkspace, error)
}
