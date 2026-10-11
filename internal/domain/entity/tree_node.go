package entity

import (
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const (
	copySuffix    = " Copy"
	errorIdPrefix = "error:"
)

type TreeNode struct {
	ID       string
	Kind     enums.ItemKind
	Name     string
	Method   string
	URL      string
	Error    string
	Warning  string
	FileName string
	Children []TreeNode
	Entries  []string
	Examples []TreeExample
	Ignored  bool
}

func NewRequestNode(request Request, fileName string) TreeNode {
	return TreeNode{
		ID:       request.ID,
		Kind:     itemkind.Request,
		Name:     request.Name,
		Method:   request.NormalizedMethod(),
		URL:      request.URL,
		FileName: fileName,
		Examples: treeExamples(request.Examples),
	}
}

func treeExamples(examples []SavedExample) []TreeExample {
	nodes := make([]TreeExample, 0, len(examples))
	for _, example := range examples {
		nodes = append(nodes, example.TreeExample())
	}
	return nodes
}

func NewContainerNode(container Collection, fileName string, entries []string) TreeNode {
	return TreeNode{
		ID:       container.ID,
		Kind:     container.Kind,
		Name:     container.Name,
		FileName: fileName,
		Entries:  entries,
	}
}

func NewErrorNode(relativePath, fileName, reason string, ignored bool) TreeNode {
	return TreeNode{
		ID:       errorIdPrefix + relativePath,
		Kind:     itemkind.Error,
		Name:     fileName,
		FileName: fileName,
		Error:    reason,
		Ignored:  ignored,
	}
}

func (t TreeNode) WithWarning(warning string) TreeNode {
	t.Warning = warning
	return t
}

func (t TreeNode) IsContainer() bool {
	return t.Kind == itemkind.Collection || t.Kind == itemkind.Folder
}

func (t TreeNode) IsCollection() bool {
	return t.Kind == itemkind.Collection
}

func (t TreeNode) IsRequest() bool {
	return t.Kind == itemkind.Request
}

func (t TreeNode) IsItem() bool {
	return t.Kind != itemkind.Error
}

func (t TreeNode) HasName(name string) bool {
	return strings.EqualFold(strings.TrimSpace(t.Name), strings.TrimSpace(name))
}

func (t TreeNode) CopyName() string {
	return NormalizeItemName(t.Name + copySuffix)
}

func (t TreeNode) CopyAs(name, fileName string, newID func() string) ItemCopy {
	copied := ItemCopy{ID: newID(), Name: name, FileName: fileName, Kind: t.Kind, Children: []ItemCopy{}}
	for _, child := range t.Children {
		if !child.Ignored {
			copied.Children = append(copied.Children, child.CopyAs("", child.FileName, newID))
		}
	}
	return copied
}

func (t TreeNode) Matches(query string) bool {
	needle := strings.ToLower(strings.TrimSpace(query))
	if !t.IsRequest() || needle == "" {
		return false
	}
	return strings.Contains(strings.ToLower(t.Name), needle) || strings.Contains(strings.ToLower(t.URL), needle)
}

func (t TreeNode) RequestIDs() []string {
	ids := make([]string, 0)
	if t.IsRequest() {
		ids = append(ids, t.ID)
	}
	for _, child := range t.Children {
		ids = append(ids, child.RequestIDs()...)
	}
	return ids
}

func (t TreeNode) ChildOrder() []string {
	order := make([]string, 0, len(t.Children))
	for _, child := range t.Children {
		if !child.Ignored {
			order = append(order, child.FileName)
		}
	}
	return order
}

func (t TreeNode) OrderAppending(fileName string) []string {
	return append(t.ChildOrder(), fileName)
}

func (t TreeNode) OrderInsertingAfter(existing, fileName string) []string {
	order := make([]string, 0, len(t.Children)+1)
	for _, name := range t.ChildOrder() {
		order = append(order, name)
		if name == existing {
			order = append(order, fileName)
		}
	}
	return order
}

func (t TreeNode) OrderRenaming(previous, fileName string) []string {
	order := t.ChildOrder()
	for index, name := range order {
		if name == previous {
			order[index] = fileName
		}
	}
	return order
}

func (t TreeNode) OrderRemoving(fileName string) []string {
	order := make([]string, 0, len(t.Children))
	for _, name := range t.ChildOrder() {
		if name != fileName {
			order = append(order, name)
		}
	}
	return order
}

func (t TreeNode) OrderImporting(replacements map[string]string, appended []string) []string {
	order := make([]string, 0, len(t.Children)+len(appended))
	for _, name := range t.ChildOrder() {
		if replacement, found := replacements[name]; found {
			name = replacement
		}
		order = append(order, name)
	}
	return append(order, appended...)
}

func (t TreeNode) EntriesExcept(fileName string) []string {
	return t.EntriesWithout([]string{fileName})
}

func (t TreeNode) EntriesReplacing(replaced []TreeNode) []string {
	fileNames := make([]string, 0, len(replaced))
	for _, node := range replaced {
		fileNames = append(fileNames, node.FileName)
	}
	return t.EntriesWithout(fileNames)
}

func (t TreeNode) EntriesWithout(fileNames []string) []string {
	entries := make([]string, 0, len(t.Entries))
	for _, entry := range t.Entries {
		if !slices.Contains(fileNames, entry) {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (t TreeNode) find(id string) (*TreeNode, *TreeNode) {
	for index := range t.Children {
		child := &t.Children[index]
		if child.ID == id && child.IsItem() {
			return child, &t
		}
		if found, parent := child.find(id); found != nil {
			return found, parent
		}
	}
	return nil, nil
}

func NamedTreeNode(nodes []TreeNode, name string) (TreeNode, bool) {
	for _, node := range nodes {
		if node.IsItem() && node.HasName(name) {
			return node, true
		}
	}
	return TreeNode{}, false
}

func (t TreeNode) Contains(id string) bool {
	if t.ID == id {
		return true
	}
	for _, child := range t.Children {
		if child.IsItem() && child.Contains(id) {
			return true
		}
	}
	return false
}

func (t TreeNode) RenamesOnMove(item, source TreeNode) bool {
	return source.ID != t.ID && t.HoldsEntry(item.FileName)
}

func (t TreeNode) HoldsEntry(fileName string) bool {
	for _, entry := range t.Entries {
		if strings.EqualFold(entry, fileName) {
			return true
		}
	}
	return false
}

func (t TreeNode) OrderInserting(fileName string, index int) []string {
	return insertAt(t.ChildOrder(), fileName, index)
}

func (t TreeNode) OrderMoving(fileName string, index int) []string {
	return insertAt(t.OrderRemoving(fileName), fileName, index)
}

func insertAt(order []string, fileName string, index int) []string {
	index = min(max(index, 0), len(order))
	return slices.Insert(order, index, fileName)
}
