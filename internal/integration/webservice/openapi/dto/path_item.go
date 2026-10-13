package dto

import (
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

const parametersKey = "parameters"

var methods = []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}

type PathItem struct {
	Path       string
	Parameters []Parameter
	Operations []PathOperation
}

func IsMethod(key string) bool {
	return slices.Contains(methods, strings.ToLower(key))
}

func (p PathItem) MarshalYAML() (any, error) {
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: mapTag}
	for _, operation := range p.Operations {
		value := &yaml.Node{}
		if err := value.Encode(operation.Operation); err != nil {
			return nil, err
		}
		mapping.Content = append(mapping.Content, keyNode(strings.ToLower(operation.Method)), value)
	}
	return mapping, nil
}

func (p *PathItem) UnmarshalYAML(node *yaml.Node) error {
	node = resolved(node)
	for index := 0; index+1 < len(node.Content); index += 2 {
		key, value := node.Content[index].Value, resolved(node.Content[index+1])
		switch {
		case key == parametersKey:
			if err := value.Decode(&p.Parameters); err != nil {
				return err
			}
		case IsMethod(key):
			var operation Operation
			if err := value.Decode(&operation); err != nil {
				return err
			}
			p.Operations = append(p.Operations, PathOperation{Method: key, Operation: operation})
		}
	}
	return nil
}
