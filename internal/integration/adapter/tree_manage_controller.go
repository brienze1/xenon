package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type TreeManageController interface {
	CreateCollection(ctx context.Context, request dto.CreateCollection) (*dto.TreeNode, error)
	CreateFolder(ctx context.Context, request dto.CreateFolder) (*dto.TreeNode, error)
	CreateRequest(ctx context.Context, request dto.CreateRequest) (*dto.TreeNode, error)
	Rename(ctx context.Context, request dto.RenameItem) (*dto.TreeNode, error)
	Duplicate(ctx context.Context, request dto.TreeItemReference) (*dto.TreeNode, error)
	Delete(ctx context.Context, request dto.TreeItemReference) (*dto.DeletedItems, error)
}
