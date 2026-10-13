package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type DuplicateTreeItem interface {
	Duplicate(ctx context.Context, root, id string, copied entity.ItemCopy) (*entity.TreeNode, error)
}
