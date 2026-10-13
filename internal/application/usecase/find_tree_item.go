package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindTreeItem interface {
	FindNode(ctx context.Context, root, id string) (*entity.TreeNode, error)
	FindContainer(ctx context.Context, root, id string) (*entity.Collection, error)
	FindAncestors(ctx context.Context, root, id string) ([]entity.Collection, error)
}
