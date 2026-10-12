package entity

type ChangeSet struct {
	Items []ChangeItem
}

func NewChangeSet(items []ChangeItem) ChangeSet {
	return ChangeSet{Items: append([]ChangeItem{}, items...)}
}

func (c ChangeSet) Selected(ids []string) ChangeSet {
	selected := make([]ChangeItem, 0, len(ids))
	for _, item := range c.Items {
		if item.IsAnyOf(ids) {
			selected = append(selected, item)
		}
	}
	return ChangeSet{Items: selected}
}

func (c ChangeSet) IsEmpty() bool {
	return len(c.Items) == 0
}

func (c ChangeSet) Count() int {
	return len(c.Items)
}

func (c ChangeSet) Paths() []string {
	paths := make([]string, 0, len(c.Items))
	for _, item := range c.Items {
		paths = append(paths, item.Paths...)
	}
	return paths
}

func (c ChangeSet) TreeFrom(head map[string]string, worktree SyncedFiles) map[string]string {
	tree := make(map[string]string, len(head))
	for path, hash := range head {
		tree[path] = hash
	}
	for _, path := range c.Paths() {
		if hash, present := worktree[path]; present {
			tree[path] = hash
			continue
		}
		delete(tree, path)
	}
	return tree
}

func (c ChangeSet) PresentPaths(worktree SyncedFiles) []string {
	present := make([]string, 0)
	for _, path := range c.Paths() {
		if _, found := worktree[path]; found {
			present = append(present, path)
		}
	}
	return present
}
