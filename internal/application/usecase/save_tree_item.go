package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveTreeItem interface {
	CreateContainer(ctx context.Context, root, parentId, fileName string, container entity.Collection) (*entity.TreeNode, error)
	CreateRequest(ctx context.Context, root, parentId, fileName string, request entity.Request) (*entity.Request, error)
	Rename(ctx context.Context, root, id, name, fileName string) (*entity.TreeNode, error)
	Duplicate(ctx context.Context, root, id, name, fileName string) (*entity.TreeNode, error)
	SaveContainer(ctx context.Context, root string, container entity.Collection) (*entity.Collection, error)
}
