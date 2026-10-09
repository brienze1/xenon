package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func SearchTree(tree entity.Tree, query string) entity.TreeSearch {
	result := entity.NewTreeSearch()
	searchNodes(tree.Collections(), []entity.TreeNode{}, query, &result)
	return result
}

func searchNodes(nodes, ancestors []entity.TreeNode, query string, result *entity.TreeSearch) {
	for _, node := range nodes {
		if node.Matches(query) {
			result.AddMatch(node.ID, ancestors)
		}
		searchNodes(node.Children, append(append([]entity.TreeNode{}, ancestors...), node), query, result)
	}
}
