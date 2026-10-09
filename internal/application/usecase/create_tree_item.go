package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type CreateTreeItem interface {
	CreateContainer(ctx context.Context, root, parentId, fileName string, container entity.Collection) (*entity.TreeNode, error)
	CreateRequest(ctx context.Context, root, parentId, fileName string, request entity.Request) (*entity.Request, error)
}
