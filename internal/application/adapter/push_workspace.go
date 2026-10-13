package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PushWorkspaceService interface {
	Push(ctx context.Context, localId string) (*entity.SyncStatus, error)
}
