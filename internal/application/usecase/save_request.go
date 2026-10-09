package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveRequest interface {
	Save(ctx context.Context, root string, request entity.Request, baseRevision string, force bool) (*entity.Request, error)
}
