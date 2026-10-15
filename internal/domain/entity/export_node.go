package entity

import "strings"

type ExportNode struct {
	Request          *Request
	Path             string
	Children         []ExportNode
	InheritedScripts []Scripts
	Container        Collection
}

func NewRequestExportNode(request Request, path string) ExportNode {
	return ExportNode{Request: &request, Path: path, Children: []ExportNode{}, InheritedScripts: []Scripts{}}
}

func NewContainerExportNode(container Collection, path string, children []ExportNode) ExportNode {
	return ExportNode{Container: container, Path: path, Children: children, InheritedScripts: []Scripts{}}
}

func (e ExportNode) IsRequest() bool {
	return e.Request != nil
}

func (e ExportNode) ID() string {
	if e.IsRequest() {
		return e.Request.ID
	}
	return e.Container.ID
}

func (e ExportNode) Auth() Auth {
	if e.IsRequest() {
		return e.Request.Auth
	}
	return e.Container.Auth
}

func (e ExportNode) HasContainerHeaders() bool {
	return !e.IsRequest() && len(e.Container.Headers) > 0
}

func (e ExportNode) ScopeRefWithin(ancestors []Collection) ScopeRef {
	if collection, found := RootCollection(ancestors); found {
		return collection.ScopeRef()
	}
	return e.Container.ScopeRef()
}

func (e ExportNode) Name() string {
	if e.IsRequest() {
		return e.Request.Name
	}
	return e.Container.Name
}

func (e ExportNode) Rooted(container Collection, inherited []Scripts) ExportNode {
	e.Container = container
	e.InheritedScripts = append([]Scripts{}, inherited...)
	return e
}

func (e ExportNode) WithChildren(children []ExportNode) ExportNode {
	e.Children = children
	return e
}

func (e ExportNode) Counts() ExportCounts {
	if e.IsRequest() {
		return ExportCounts{Requests: 1, Examples: len(e.Request.Examples)}
	}
	counts := ExportCounts{Variables: len(e.Container.OwnVariables())}
	if e.Container.IsFolder() {
		counts.Folders = 1
	} else {
		counts.Collections = 1
	}
	for _, child := range e.Children {
		counts = counts.Add(child.Counts())
	}
	return counts
}

func (e ExportNode) Nodes() []ExportNode {
	nodes := []ExportNode{e}
	for _, child := range e.Children {
		nodes = append(nodes, child.Nodes()...)
	}
	return nodes
}

func ExportPath(ancestors []Collection, name string) string {
	path := ""
	for _, ancestor := range ancestors {
		path = JoinImportPath(path, ancestor.Name)
	}
	return JoinImportPath(path, name)
}

func ExportableChildren(node TreeNode) []TreeNode {
	children := make([]TreeNode, 0, len(node.Children))
	for _, child := range node.Children {
		if child.IsItem() && !child.Ignored {
			children = append(children, child)
		}
	}
	return children
}

// WithAliasesAsRequests writes every request's aliases as requests right after it, since Postman has no aliases.
func (e ExportNode) WithAliasesAsRequests() ExportNode {
	if e.IsRequest() {
		request := *e.Request
		request.Aliases = nil
		e.Request = &request
		return e
	}
	children := make([]ExportNode, 0, len(e.Children))
	for _, child := range e.Children {
		children = append(children, child.WithAliasRequests()...)
	}
	e.Children = children
	return e
}

func (e ExportNode) WithAliasRequests() []ExportNode {
	return append([]ExportNode{e.WithAliasesAsRequests()}, e.AliasRequests()...)
}

// AliasRequests are a request's aliases as requests named "<request> (<alias>)", without examples.
func (e ExportNode) AliasRequests() []ExportNode {
	if !e.IsRequest() {
		return []ExportNode{}
	}
	nodes := make([]ExportNode, 0, len(e.Request.Aliases))
	parent := strings.TrimSuffix(e.Path, e.Request.Name)
	for _, alias := range e.Request.Aliases {
		version := e.Request.Version(alias.ID)
		name := version.Name + " (" + alias.Name + ")"
		version = version.Identified(version.ID+"-"+alias.ID, name)
		version.Examples = nil
		version.Aliases = nil
		nodes = append(nodes, ExportNode{Request: &version, Path: parent + name, Children: []ExportNode{}, InheritedScripts: e.InheritedScripts})
	}
	return nodes
}
