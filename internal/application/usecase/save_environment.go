package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveEnvironment interface {
	Save(ctx context.Context, root string, environment entity.Environment) (*entity.Environment, error)
}
