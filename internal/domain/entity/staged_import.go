package entity

type StagedImport struct {
	Parent     *Collection
	Collection *Collection
	SourceID   string
	ItemKey    string
	Root       string
	ParentID   string
	Dir        string
	Nodes      []ImportNode
	Order      []string
	Replaces   []string
	Merge      bool
}

func (s StagedImport) WithDir(dir string) StagedImport {
	s.Dir = dir
	return s
}

func (s StagedImport) LogValue() any {
	return map[string]any{"sourceId": s.SourceID, "itemKey": s.ItemKey, "parentId": s.ParentID, "merge": s.Merge, "replaces": len(s.Replaces)}
}
