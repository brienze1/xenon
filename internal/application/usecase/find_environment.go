package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindEnvironment interface {
	Find(ctx context.Context, root, id string) (*entity.Environment, error)
}
