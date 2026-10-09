package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindRequest interface {
	Find(ctx context.Context, root, id string) (*entity.Request, error)
	FindRevision(ctx context.Context, root, id string) (*entity.RequestChange, error)
}
