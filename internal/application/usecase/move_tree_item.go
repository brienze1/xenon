package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type MoveTreeItem interface {
	Move(ctx context.Context, root string, move entity.TreeMove) (*entity.Tree, error)
}
