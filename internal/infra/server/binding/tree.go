package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type TreeService struct {
	search   adapter.TreeSearchController
	manage   adapter.TreeManageController
	settings adapter.TreeSettingsController
}

func NewTreeService() *TreeService {
	return &TreeService{
		search:   controller.GetTreeSearchController(),
		manage:   controller.GetTreeManageController(),
		settings: controller.GetTreeSettingsController(),
	}
}

func (t *TreeService) GetTree(ctx context.Context, request dto.GetTree) (dto.TreeNodeList, error) {
	return t.search.GetTree(ctx, request)
}

func (t *TreeService) Search(ctx context.Context, request dto.SearchTree) (*dto.TreeSearch, error) {
	return t.search.Search(ctx, request)
}

func (t *TreeService) CreateCollection(ctx context.Context, request dto.CreateCollection) (*dto.TreeNode, error) {
	return t.manage.CreateCollection(ctx, request)
}

func (t *TreeService) CreateFolder(ctx context.Context, request dto.CreateFolder) (*dto.TreeNode, error) {
	return t.manage.CreateFolder(ctx, request)
}

func (t *TreeService) CreateRequest(ctx context.Context, request dto.CreateRequest) (*dto.TreeNode, error) {
	return t.manage.CreateRequest(ctx, request)
}

func (t *TreeService) Rename(ctx context.Context, request dto.RenameItem) (*dto.TreeNode, error) {
	return t.manage.Rename(ctx, request)
}

func (t *TreeService) Duplicate(ctx context.Context, request dto.TreeItemReference) (*dto.TreeNode, error) {
	return t.manage.Duplicate(ctx, request)
}

func (t *TreeService) Delete(ctx context.Context, request dto.TreeItemReference) (*dto.DeletedItems, error) {
	return t.manage.Delete(ctx, request)
}

func (t *TreeService) GetSettings(ctx context.Context, request dto.TreeItemReference) (*dto.ItemSettings, error) {
	return t.settings.GetSettings(ctx, request)
}

func (t *TreeService) SaveSettings(ctx context.Context, request dto.SaveItemSettings) (*dto.ItemSettings, error) {
	return t.settings.SaveSettings(ctx, request)
}
