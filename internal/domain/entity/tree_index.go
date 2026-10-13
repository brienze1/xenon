package entity

import "sort"

type TreeIndex struct {
	Items        map[string]IndexedItem
	Reidentified map[string]string
	Tree         Tree
}

func NewTreeIndex(tree Tree, items map[string]IndexedItem, reidentified map[string]string) TreeIndex {
	return TreeIndex{Tree: tree, Items: items, Reidentified: reidentified}
}

func (t TreeIndex) Item(id string) (IndexedItem, bool) {
	item, found := t.Items[id]
	return item, found
}

func (t TreeIndex) Size() int {
	return len(t.Items)
}

func (t TreeIndex) RequestIDsInside(dir string) []string {
	ids := make([]string, 0)
	for id, item := range t.Items {
		if item.IsRequest() && item.IsInside(dir) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}
