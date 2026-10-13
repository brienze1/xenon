package dto

import "go.yaml.in/yaml/v3"

// Value keeps an example exactly as written, whatever its YAML shape.
type Value struct {
	node *yaml.Node
}

func NewValue(node *yaml.Node) *Value {
	return &Value{node: node}
}

func (v *Value) MarshalYAML() (any, error) {
	return v.node, nil
}

func (v *Value) UnmarshalYAML(node *yaml.Node) error {
	v.node = node
	return nil
}

func (v *Value) Node() *yaml.Node {
	if v == nil {
		return nil
	}
	return v.node
}
