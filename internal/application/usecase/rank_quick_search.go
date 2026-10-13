package usecase

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func RankQuickSearch(tree entity.Tree, query string, limit int) []entity.QuickSearchHit {
	hits := make([]entity.QuickSearchHit, 0)
	collectQuickSearchHits(tree.Collections(), []entity.TreeNode{}, query, &hits)
	slices.SortStableFunc(hits, func(a, b entity.QuickSearchHit) int {
		return a.Rank - b.Rank
	})
	return hits[:min(len(hits), entity.QuickSearchLimit(limit))]
}

func collectQuickSearchHits(nodes, ancestors []entity.TreeNode, query string, hits *[]entity.QuickSearchHit) {
	for _, node := range nodes {
		if hit, found := entity.NewQuickSearchHit(node, ancestors, query); found {
			*hits = append(*hits, hit)
		}
		if node.IsContainer() {
			collectQuickSearchHits(node.Children, append(slices.Clone(ancestors), node), query, hits)
		}
	}
}
