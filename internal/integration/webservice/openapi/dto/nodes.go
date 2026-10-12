package dto

import (
	"go.yaml.in/yaml/v3"
)

const (
	stringTag = "!!str"
	mapTag    = "!!map"
)

func keyNode(key string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: stringTag, Value: key}
}

func emptyMapping() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: mapTag, Style: yaml.FlowStyle}
}

func mappingOf[T any](entries []T, name func(T) string) (*yaml.Node, error) {
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: mapTag}
	for _, entry := range entries {
		value := &yaml.Node{}
		if err := value.Encode(entry); err != nil {
			return nil, err
		}
		mapping.Content = append(mapping.Content, keyNode(name(entry)), value)
	}
	if len(mapping.Content) == 0 {
		return emptyMapping(), nil
	}
	return mapping, nil
}

func entriesOf[T any](node *yaml.Node, named func(*T, string)) ([]T, error) {
	node = resolved(node)
	entries := make([]T, 0, len(node.Content)/2)
	if node.Kind != yaml.MappingNode {
		return entries, nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		var entry T
		if err := resolved(node.Content[index+1]).Decode(&entry); err != nil {
			return nil, err
		}
		named(&entry, node.Content[index].Value)
		entries = append(entries, entry)
	}
	return entries, nil
}

func resolved(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	if node != nil && node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return resolved(node.Content[0])
	}
	return node
}

type field struct {
	value any
	key   string
	keep  bool
}

func always(key string, value any) field {
	return field{key: key, value: value, keep: true}
}

func text(key, value string) field {
	return field{key: key, value: value, keep: value != ""}
}

func when(present bool, key string, value any) field {
	return field{key: key, value: value, keep: present}
}

// mappingOfFields writes the fields in the order given, so the document reads the same every time.
func mappingOfFields(fields ...field) (*yaml.Node, error) {
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: mapTag}
	for _, entry := range fields {
		if !entry.keep {
			continue
		}
		value := &yaml.Node{}
		if err := value.Encode(entry.value); err != nil {
			return nil, err
		}
		mapping.Content = append(mapping.Content, keyNode(entry.key), value)
	}
	return mapping, nil
}
