package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ListEnvironments interface {
	List(ctx context.Context, root string) ([]entity.Environment, error)
}
