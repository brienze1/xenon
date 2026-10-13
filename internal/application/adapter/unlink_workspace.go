package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type UnlinkWorkspaceService interface {
	Unlink(ctx context.Context, localId, targetFolder string) (*entity.RegisteredWorkspace, error)
}
