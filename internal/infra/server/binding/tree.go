package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type TreeService struct {
	search        adapter.TreeSearchController
	manage        adapter.TreeManageController
	settings      adapter.TreeSettingsController
	documentation adapter.TreeDocumentationController
}

func NewTreeService(
	search adapter.TreeSearchController,
	manage adapter.TreeManageController,
	settings adapter.TreeSettingsController,
	documentation adapter.TreeDocumentationController,
) *TreeService {
	return &TreeService{search: search, manage: manage, settings: settings, documentation: documentation}
}

func (t *TreeService) GetTree(ctx context.Context, request dto.GetTree) (dto.TreeNodeList, error) {
	return t.search.GetTree(ctx, request)
}

func (t *TreeService) Search(ctx context.Context, request dto.SearchTree) (*dto.TreeSearch, error) {
	return t.search.Search(ctx, request)
}

func (t *TreeService) QuickSearch(ctx context.Context, request dto.QuickSearch) (dto.QuickSearchHitList, error) {
	return t.search.QuickSearch(ctx, request)
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

func (t *TreeService) Move(ctx context.Context, request dto.MoveItem) (dto.TreeNodeList, error) {
	return t.manage.Move(ctx, request)
}

func (t *TreeService) GetSettings(ctx context.Context, request dto.TreeItemReference) (*dto.ItemSettings, error) {
	return t.settings.GetSettings(ctx, request)
}

func (t *TreeService) SaveSettings(ctx context.Context, request dto.SaveItemSettings) (*dto.ItemSettings, error) {
	return t.settings.SaveSettings(ctx, request)
}

func (t *TreeService) GetDocumentation(ctx context.Context, request dto.TreeItemReference) (*dto.Documentation, error) {
	return t.documentation.GetDocumentation(ctx, request)
}

func (t *TreeService) PreviewDocumentation(ctx context.Context, request dto.DocumentationInput) (*dto.Documentation, error) {
	return t.documentation.PreviewDocumentation(ctx, request)
}

func (t *TreeService) PlanDocumentation(ctx context.Context, request dto.DocumentationInput) (*dto.DocumentationPlan, error) {
	return t.documentation.PlanDocumentation(ctx, request)
}

func (t *TreeService) ApplyDocumentation(ctx context.Context, request dto.DocumentationInput) (*dto.DocumentationApplied, error) {
	return t.documentation.ApplyDocumentation(ctx, request)
}
