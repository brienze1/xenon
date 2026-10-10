package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const importPathSeparator = "/"

type ImportNode struct {
	Request   *Request
	FileName  string
	Children  []ImportNode
	Container Collection
}

func NewRequestImportNode(request Request) ImportNode {
	return ImportNode{Request: &request, Children: []ImportNode{}}
}

func NewContainerImportNode(container Collection, children []ImportNode) ImportNode {
	return ImportNode{Container: container, Children: children}
}

func (i ImportNode) IsRequest() bool {
	return i.Request != nil
}

func (i ImportNode) ID() string {
	if i.IsRequest() {
		return i.Request.ID
	}
	return i.Container.ID
}

func (i ImportNode) Name() string {
	if i.IsRequest() {
		return i.Request.Name
	}
	return i.Container.Name
}

func (i ImportNode) Kind() enums.ItemKind {
	if i.IsRequest() {
		return itemkind.Request
	}
	return i.Container.Kind
}

func (i ImportNode) Renamed(name string) ImportNode {
	if i.IsRequest() {
		request := *i.Request
		request.Name = name
		i.Request = &request
		return i
	}
	i.Container.Name = name
	return i
}

func (i ImportNode) AsFolder() ImportNode {
	i.Container.Kind = itemkind.Folder
	i.Container.Variables = []VariableRow{}
	return i
}

func (i ImportNode) WithFileName(fileName string) ImportNode {
	i.FileName = fileName
	return i
}

func (i ImportNode) WithChildren(children []ImportNode) ImportNode {
	i.Children = children
	order := make([]string, 0, len(children))
	for _, child := range children {
		order = append(order, child.FileName)
	}
	i.Container.Order = order
	return i
}

func (i ImportNode) Counts() ImportCounts {
	if i.IsRequest() {
		return ImportCounts{Requests: 1, Examples: len(i.Request.Examples), Scripts: len(i.Request.Scripts.Phases())}
	}
	counts := ImportCounts{Scripts: len(i.Container.Scripts.Phases())}
	if i.Container.IsFolder() {
		counts.Folders = 1
	} else {
		counts.Collections = 1
		counts.Variables = len(i.Container.Variables)
	}
	for _, child := range i.Children {
		counts = counts.Add(child.Counts())
	}
	return counts
}

func (i ImportNode) Scripts(parentPath string) []ScriptRef {
	path := JoinImportPath(parentPath, i.Name())
	scripts := i.Container.Scripts
	if i.IsRequest() {
		scripts = i.Request.Scripts
	}
	refs := make([]ScriptRef, 0)
	for _, phase := range scripts.Phases() {
		refs = append(refs, ScriptRef{ItemID: i.ID(), Path: path, Phase: phase, Text: scripts.Text(phase)})
	}
	for _, child := range i.Children {
		refs = append(refs, child.Scripts(path)...)
	}
	return refs
}

func (i ImportNode) PlacedInto(parent Collection) ImportPlacement {
	if i.IsRequest() {
		return ImportPlacement{Parent: &parent, Nodes: []ImportNode{i}}
	}
	if parent.HasOwnSettings() {
		return ImportPlacement{Parent: &parent, Nodes: []ImportNode{i.AsFolder()}, Wrapped: true}
	}
	merged := parent.WithImportedSettings(i.Container)
	return ImportPlacement{Parent: &merged, Nodes: i.Children, Merge: true}
}

func (i ImportNode) WithScripts(scripts map[string]Scripts) ImportNode {
	if translated, found := scripts[i.ID()]; found && i.IsRequest() {
		request := *i.Request
		request.Scripts = translated
		i.Request = &request
	} else if found {
		i.Container.Scripts = translated
	}
	children := make([]ImportNode, 0, len(i.Children))
	for _, child := range i.Children {
		children = append(children, child.WithScripts(scripts))
	}
	i.Children = children
	return i
}

func (i ImportNode) IDs() []string {
	ids := []string{i.ID()}
	for _, child := range i.Children {
		ids = append(ids, child.IDs()...)
	}
	return ids
}

func (i ImportNode) ChildNames() []string {
	names := make([]string, 0, len(i.Children))
	for _, child := range i.Children {
		names = append(names, child.Name())
	}
	return names
}

func JoinImportPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return strings.Join([]string{parent, name}, importPathSeparator)
}
