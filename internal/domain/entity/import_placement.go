package entity

type ImportPlacement struct {
	Parent  *Collection
	Nodes   []ImportNode
	Merge   bool
	Wrapped bool
}

func NewTopLevelPlacement(root ImportNode) ImportPlacement {
	return ImportPlacement{Nodes: []ImportNode{root.AsCollection()}}
}

func (i ImportPlacement) HasParent() bool {
	return i.Parent != nil
}

func (i ImportPlacement) WithNodes(nodes []ImportNode) ImportPlacement {
	i.Nodes = nodes
	return i
}
