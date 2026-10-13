package yamlnode

import (
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Field struct {
	Value        *yaml.Node
	Key          string
	Fields       []Field
	KeepExisting bool
	Remove       bool
}

func String(key, value string) Field {
	return Field{Key: key, Value: StringNode(value)}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(value)}}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(value)}}
}

func StringList(key string, values []string) Field {
	items := make([]*yaml.Node, 0, len(values))
	for _, value := range values {
		items = append(items, StringNode(value))
	}
	return List(key, items)
}

func List(key string, items []*yaml.Node) Field {
	sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: items}
	if len(items) == 0 {
		sequence.Style = yaml.FlowStyle
	}
	return Field{Key: key, Value: sequence}
}

func Object(key string, fields ...Field) Field {
	return Field{Key: key, Fields: append([]Field{}, fields...)}
}

func Reserved(key string, value *yaml.Node) Field {
	return Field{Key: key, Value: value, KeepExisting: true}
}

func Absent(key string) Field {
	return Field{Key: key, Remove: true}
}

func StringNode(value string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
	if strings.Contains(value, "\n") {
		node.Style = yaml.LiteralStyle
	}
	return node
}

func EmptyMapping() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Style: yaml.FlowStyle}
}

func EmptyList() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
}

func MappingNode(fields ...Field) *yaml.Node {
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	Patch(mapping, fields)
	if len(mapping.Content) == 0 {
		mapping.Style = yaml.FlowStyle
	}
	return mapping
}

func (f Field) node() *yaml.Node {
	if f.Fields != nil {
		return MappingNode(f.Fields...)
	}
	return f.Value
}

// Any encodes a decoded YAML or JSON value as a node, empty maps and lists in flow style; panics if it cannot be encoded.
func Any(value any) *yaml.Node {
	node := &yaml.Node{}
	if err := node.Encode(value); err != nil {
		panic(err)
	}
	if (node.Kind == yaml.MappingNode || node.Kind == yaml.SequenceNode) && len(node.Content) == 0 {
		node.Style = yaml.FlowStyle
	}
	return node
}
