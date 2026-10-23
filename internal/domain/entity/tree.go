package entity

import "github.com/brienze1/xenon/internal/domain/enums/itemkind"

type Tree struct {
	Root TreeNode
}

func NewTree(collections []TreeNode, entries []string) Tree {
	return Tree{Root: TreeNode{Kind: itemkind.Workspace, Children: collections, Entries: entries}}
}

func (t Tree) Collections() []TreeNode {
	return t.Root.Children
}

func (t Tree) Item(id string) (*TreeNode, bool) {
	item, _ := t.Root.find(id)
	return item, item != nil
}

func (t Tree) Parent(id string) (*TreeNode, bool) {
	_, parent := t.Root.find(id)
	return parent, parent != nil
}

func (t Tree) Container(id string) (*TreeNode, bool) {
	item, found := t.Item(id)
	if !found || !item.IsContainer() {
		return nil, false
	}
	return item, true
}

func (t Tree) SiblingEntries(id string) []string {
	item, found := t.Item(id)
	if !found {
		return []string{}
	}
	parent, _ := t.Parent(id)
	return parent.EntriesExcept(item.FileName)
}

func (t Tree) PathOf(id string) string {
	path := ""
	for _, ancestor := range t.Ancestors(id) {
		path = JoinImportPath(path, ancestor.Name)
	}
	if item, found := t.Item(id); found {
		path = JoinImportPath(path, item.Name)
	}
	return path
}

func (t Tree) Ancestors(id string) []TreeNode {
	path, _ := ancestorsOf(t.Root.Children, id)
	return path
}

func ancestorsOf(nodes []TreeNode, id string) ([]TreeNode, bool) {
	for _, node := range nodes {
		if node.ID == id && node.IsItem() {
			return []TreeNode{}, true
		}
		if below, found := ancestorsOf(node.Children, id); found {
			return append([]TreeNode{node}, below...), true
		}
	}
	return nil, false
}

// ContainerIDs lists the id of every collection and folder in the tree.
func (t Tree) ContainerIDs() []string {
	return containerIDs(t.Root.Children)
}

func containerIDs(nodes []TreeNode) []string {
	ids := make([]string, 0)
	for _, node := range nodes {
		if node.IsContainer() && node.IsItem() {
			ids = append(ids, node.ID)
			ids = append(ids, containerIDs(node.Children)...)
		}
	}
	return ids
}
