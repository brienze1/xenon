package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PullWorkspaceService interface {
	Pull(ctx context.Context, localId string) (*entity.PullResult, error)
}
