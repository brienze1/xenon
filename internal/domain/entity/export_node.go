package entity

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
