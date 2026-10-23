package entity

import "strings"

// Entries lists every node then its examples then its children, in tree order.
func (t Tree) Entries() []TreeEntry {
	return appendTreeEntries(make([]TreeEntry, 0), t.Root.Children, "", "")
}

func appendTreeEntries(entries []TreeEntry, nodes []TreeNode, parentID, parentPath string) []TreeEntry {
	for _, node := range nodes {
		path := joinAIItemPath(parentPath, node.Name)
		entries = append(entries, TreeEntry{Node: node, ParentID: parentID, Path: path})
		for _, example := range node.Examples {
			entries = append(entries, TreeEntry{Node: node, Example: &example, RequestID: node.ID, ParentID: node.ID, Path: joinAIItemPath(path, example.Name)})
		}
		entries = appendTreeEntries(entries, node.Children, node.ID, path)
	}
	return entries
}

func (t Tree) Entry(id string) (TreeEntry, bool) {
	for _, entry := range t.Entries() {
		if entry.ID() == id {
			return entry, true
		}
	}
	return TreeEntry{}, false
}

// Search keeps tree order: the matched ids plus every entry whose name contains the query, at most limit.
func (t Tree) Search(query string, matchIDs []string, limit int) []TreeEntry {
	matched := make(map[string]bool, len(matchIDs))
	for _, id := range matchIDs {
		matched[id] = true
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	all := t.Entries()
	found := make([]TreeEntry, 0, min(len(all), limit))
	for _, entry := range all {
		if len(found) == limit {
			break
		}
		if matched[entry.ID()] || strings.Contains(strings.ToLower(entry.Name()), needle) {
			found = append(found, entry)
		}
	}
	return found
}

// ContainerAt finds a collection or folder by its path; "/" separates the names, with or without spaces around it.
func (t Tree) ContainerAt(path string) (TreeEntry, bool) {
	wanted := normalizedItemPath(path)
	for _, entry := range t.Entries() {
		if entry.IsContainer() && normalizedItemPath(entry.Path) == wanted {
			return entry, true
		}
	}
	return TreeEntry{}, false
}

// ResolveRunTarget finds the container by itemId, or by path when itemId is empty; an unmatched value is used as the item id.
func (t Tree) ResolveRunTarget(path, itemId string) (string, AIActivityTarget) {
	if itemId == "" {
		if entry, found := t.ContainerAt(path); found {
			return entry.ID(), entry.Target()
		}
		return path, NewAIItemTarget(path)
	}
	if entry, found := t.Entry(itemId); found {
		return entry.ID(), entry.Target()
	}
	return itemId, NewAIItemTarget(itemId)
}

func normalizedItemPath(path string) string {
	names := strings.Split(path, "/")
	for index, name := range names {
		names[index] = strings.TrimSpace(name)
	}
	return strings.Join(names, "/")
}
