package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const AIItemPathSeparator = " / "

type TreeEntry struct {
	Example   *TreeExample
	RequestID string
	ParentID  string
	Path      string
	Node      TreeNode
}

func (t TreeEntry) ID() string {
	if t.IsExample() {
		return t.Example.ID
	}
	return t.Node.ID
}

func (t TreeEntry) Name() string {
	if t.IsExample() {
		return t.Example.Name
	}
	return t.Node.Name
}

func (t TreeEntry) Kind() enums.ItemKind {
	if t.IsExample() {
		return itemkind.Example
	}
	return t.Node.Kind
}

func (t TreeEntry) IsExample() bool {
	return t.Example != nil
}

func (t TreeEntry) IsRequest() bool {
	return !t.IsExample() && t.Node.IsRequest()
}

func (t TreeEntry) IsContainer() bool {
	return !t.IsExample() && t.Node.IsContainer()
}

func (t TreeEntry) IsCollection() bool {
	return !t.IsExample() && t.Node.IsCollection()
}

func (t TreeEntry) Children() int {
	if t.IsExample() {
		return 0
	}
	return len(t.Node.Children) + len(t.Node.Examples)
}

func (t TreeEntry) Method() string {
	if !t.IsRequest() {
		return ""
	}
	return t.Node.Method
}

func (t TreeEntry) Target() AIActivityTarget {
	return AIActivityTarget{ID: t.ID(), Name: t.Name(), Path: t.Path}
}

func joinAIItemPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + AIItemPathSeparator + name
}
