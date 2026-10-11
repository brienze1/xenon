package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SearchTreeService interface {
	Tree(ctx context.Context, localId string) (*entity.Tree, error)
	Search(ctx context.Context, localId, query string) (*entity.TreeSearch, error)
	QuickSearch(ctx context.Context, localId, query string, limit int) ([]entity.QuickSearchHit, error)
}
