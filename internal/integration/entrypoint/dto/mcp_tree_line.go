package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type MCPTreeLine struct {
	Id       string        `json:"id"`
	Kind     string        `json:"kind"`
	Name     string        `json:"name"`
	Method   string        `json:"method,omitempty"`
	Path     string        `json:"path,omitempty"`
	Items    []MCPTreeLine `json:"items,omitempty"`
	Children int           `json:"children,omitempty"`
}

// NewMCPTreeLinePage lists the nodes then the examples, expanding depth levels, and pages them after the cursor; false means an unknown cursor.
func NewMCPTreeLinePage(nodes []entity.TreeNode, examples []entity.TreeExample, depth int, cursor string) (MCPPage[MCPTreeLine], bool) {
	page, valid := entity.NewAIPage(NewMCPTreeLines(nodes, examples, depth), cursor)
	if !valid {
		return MCPPage[MCPTreeLine]{}, false
	}
	return NewMCPPageOf(page), true
}

func NewMCPTreeLines(nodes []entity.TreeNode, examples []entity.TreeExample, depth int) []MCPTreeLine {
	lines := make([]MCPTreeLine, 0, len(nodes)+len(examples))
	for _, node := range nodes {
		lines = append(lines, NewMCPTreeNodeLine(node, depth))
	}
	for _, example := range examples {
		lines = append(lines, NewMCPTreeExampleLine(example))
	}
	return lines
}

func NewMCPTreeNodeLine(node entity.TreeNode, depth int) MCPTreeLine {
	line := MCPTreeLine{Id: node.ID, Kind: node.Kind.String(), Name: node.Name, Children: len(node.Children) + len(node.Examples)}
	if node.IsRequest() {
		line.Method = node.Method
	}
	if depth == mcpTreeUnlimitedDepth || depth > mcpTreeDefaultDepth {
		line.Items = NewMCPTreeLines(node.Children, node.Examples, deeperMCPTreeDepth(depth))
	}
	return line
}

func NewMCPTreeExampleLine(example entity.TreeExample) MCPTreeLine {
	return MCPTreeLine{Id: example.ID, Kind: itemkind.Example.String(), Name: example.Name}
}

func NewMCPTreeSearchLine(entry entity.TreeEntry) MCPTreeLine {
	return MCPTreeLine{Id: entry.ID(), Kind: entry.Kind().String(), Name: entry.Name(), Method: entry.Method(), Path: entry.Path}
}

func NewMCPTreeSearchList(entries []entity.TreeEntry) MCPPage[MCPTreeLine] {
	lines := make([]MCPTreeLine, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, NewMCPTreeSearchLine(entry))
	}
	return NewMCPList(lines)
}

func deeperMCPTreeDepth(depth int) int {
	if depth == mcpTreeUnlimitedDepth {
		return depth
	}
	return depth - 1
}
