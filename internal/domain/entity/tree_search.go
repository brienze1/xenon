package entity

type TreeSearch struct {
	MatchIDs    []string
	AncestorIDs []string
	Tree        Tree
}

func NewTreeSearch() TreeSearch {
	return TreeSearch{MatchIDs: []string{}, AncestorIDs: []string{}}
}

func (t *TreeSearch) AddMatch(id string, ancestors []TreeNode) {
	t.MatchIDs = append(t.MatchIDs, id)
	for _, ancestor := range ancestors {
		if !t.hasAncestor(ancestor.ID) {
			t.AncestorIDs = append(t.AncestorIDs, ancestor.ID)
		}
	}
}

func (t TreeSearch) WithTree(tree Tree) TreeSearch {
	t.Tree = tree
	return t
}

func (t *TreeSearch) hasAncestor(id string) bool {
	for _, existing := range t.AncestorIDs {
		if existing == id {
			return true
		}
	}
	return false
}
