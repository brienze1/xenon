package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RenameTreeItem interface {
	Rename(ctx context.Context, root, id, name, fileName string) (*entity.TreeNode, error)
}
