package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type LoadTree interface {
	Load(ctx context.Context, root string) (*entity.Tree, error)
}
