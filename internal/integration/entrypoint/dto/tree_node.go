package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type TreeNode struct {
	Method   *string       `json:"method,omitempty"`
	Url      *string       `json:"url,omitempty"`
	Id       string        `json:"id"`
	Kind     string        `json:"kind"`
	Name     string        `json:"name"`
	Error    string        `json:"error,omitempty"`
	Warning  string        `json:"warning,omitempty"`
	Children []TreeNode    `json:"children,omitempty"`
	Examples []TreeExample `json:"examples,omitempty"`
}

func NewTreeNode(node entity.TreeNode) TreeNode {
	result := TreeNode{Id: node.ID, Kind: node.Kind.String(), Name: node.Name, Error: node.Error, Warning: node.Warning}
	if node.IsRequest() {
		method, url := node.Method, node.URL
		result.Method, result.Url = &method, &url
	}
	for _, example := range node.Examples {
		result.Examples = append(result.Examples, NewTreeExample(example))
	}
	for _, child := range node.Children {
		result.Children = append(result.Children, NewTreeNode(child))
	}
	return result
}
