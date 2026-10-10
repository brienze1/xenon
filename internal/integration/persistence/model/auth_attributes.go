package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func authAttributesNode(attributes []entity.AuthAttribute) *yaml.Node {
	fields := make([]yamlnode.Field, 0, len(attributes))
	for _, attribute := range attributes {
		fields = append(fields, yamlnode.Field{Key: attribute.Key, Value: yamlnode.Any(attribute.Value)})
	}
	return yamlnode.MappingNode(fields...)
}

func authAttributesToEntity(node *yaml.Node) []entity.AuthAttribute {
	var attributes []entity.AuthAttribute
	if node.Kind != yaml.MappingNode {
		return attributes
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		var value any
		_ = node.Content[index+1].Decode(&value)
		attributes = append(attributes, entity.AuthAttribute{Key: node.Content[index].Value, Value: value})
	}
	return attributes
}
