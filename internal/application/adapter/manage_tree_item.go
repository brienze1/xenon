package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageTreeItemService interface {
	CreateCollection(ctx context.Context, localId, name string) (*entity.TreeNode, error)
	CreateFolder(ctx context.Context, localId, parentId, name string) (*entity.TreeNode, error)
	CreateRequest(ctx context.Context, localId, parentId, name string, request entity.Request) (*entity.TreeNode, error)
	Rename(ctx context.Context, localId, id, name string) (*entity.TreeNode, error)
	Duplicate(ctx context.Context, localId, id string) (*entity.TreeNode, error)
	Delete(ctx context.Context, localId, id string) ([]string, error)
}
