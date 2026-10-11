package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type TreeSearchController interface {
	GetTree(ctx context.Context, request dto.GetTree) (dto.TreeNodeList, error)
	Search(ctx context.Context, request dto.SearchTree) (*dto.TreeSearch, error)
	QuickSearch(ctx context.Context, request dto.QuickSearch) (dto.QuickSearchHitList, error)
}
