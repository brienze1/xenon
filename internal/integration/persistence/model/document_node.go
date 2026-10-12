package model

import (
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/nodekind"
)

const stringTag = "!!str"

func NewDocumentNode(node *yaml.Node) *entity.DocumentNode {
	for node != nil && (node.Kind == yaml.DocumentNode || node.Kind == yaml.AliasNode) {
		if node.Kind == yaml.AliasNode {
			node = node.Alias
			continue
		}
		if len(node.Content) == 0 {
			return nil
		}
		node = node.Content[0]
	}
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.MappingNode:
		mapping := entity.NewMappingNode()
		for index := 0; index+1 < len(node.Content); index += 2 {
			mapping.Set(node.Content[index].Value, NewDocumentNode(node.Content[index+1]))
		}
		return mapping
	case yaml.SequenceNode:
		items := make([]*entity.DocumentNode, 0, len(node.Content))
		for _, item := range node.Content {
			items = append(items, NewDocumentNode(item))
		}
		return entity.NewSequenceNode(items)
	default:
		return entity.NewScalarNode(node.ShortTag(), node.Value)
	}
}

func DocumentYaml(node *entity.DocumentNode) *yaml.Node {
	if node == nil {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}
	}
	switch node.Kind {
	case nodekind.Mapping:
		mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, key := range node.Keys {
			mapping.Content = append(mapping.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: stringTag, Value: key}, DocumentYaml(node.Fields[key]))
		}
		if len(mapping.Content) == 0 {
			mapping.Style = yaml.FlowStyle
		}
		return mapping
	case nodekind.Sequence:
		sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, item := range node.Items {
			sequence.Content = append(sequence.Content, DocumentYaml(item))
		}
		if len(sequence.Content) == 0 {
			sequence.Style = yaml.FlowStyle
		}
		return sequence
	default:
		scalar := &yaml.Node{Kind: yaml.ScalarNode, Tag: node.Tag, Value: node.Value}
		if node.Tag == stringTag && strings.Contains(node.Value, "\n") {
			scalar.Style = yaml.LiteralStyle
		}
		return scalar
	}
}
