package entity

type ImportWrite struct {
	Parent         *Collection
	Collection     *Collection
	Nodes          []ImportNode
	Order          []string
	Replaced       []TreeNode
	KeptVariables  []VariableRow
	AddedVariables int
	Merge          bool
	Wrapped        bool
}

func (i ImportWrite) IsEmpty() bool {
	return len(i.Nodes) == 0 && !i.Merge
}

func (i ImportWrite) HasParent(id string) bool {
	return i.Parent != nil && i.Parent.ID == id
}

func (i ImportWrite) Staged(sourceId, itemKey, root, parentId string) StagedImport {
	replaces := make([]string, 0, len(i.Replaced))
	for _, node := range i.Replaced {
		replaces = append(replaces, node.ID)
	}
	return StagedImport{
		SourceID:   sourceId,
		ItemKey:    itemKey,
		Root:       root,
		ParentID:   parentId,
		Nodes:      i.Nodes,
		Order:      i.Order,
		Replaces:   replaces,
		Parent:     i.Parent,
		Collection: i.Collection,
		Merge:      i.Merge,
	}
}

func (i ImportWrite) IDs() []string {
	ids := make([]string, 0)
	for _, node := range i.Nodes {
		ids = append(ids, node.IDs()...)
	}
	return ids
}

func (i ImportWrite) Counts() ImportCounts {
	counts := ImportCounts{Variables: i.AddedVariables}
	for _, node := range i.Nodes {
		counts = counts.Add(node.Counts())
	}
	return counts
}

func (i ImportWrite) collectionID() string {
	if i.Collection != nil {
		return i.Collection.ID
	}
	return i.Parent.ID
}
