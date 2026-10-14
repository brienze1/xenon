package dto

func NewSearchTree(localId, query string) SearchTree {
	return SearchTree{LocalId: localId, Query: query}
}
