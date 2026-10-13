package yamlnode

import "go.yaml.in/yaml/v3"

// Patch writes fields into mapping in order, keeping unknown keys where they are.
func Patch(mapping *yaml.Node, fields []Field) {
	previous := -2
	for _, field := range fields {
		index := keyIndex(mapping, field.Key)
		if field.Remove {
			removePair(mapping, index)
			continue
		}
		if index < 0 {
			index = previous + 2
			insertPair(mapping, index, field)
			previous = index
			continue
		}
		previous = index
		if field.KeepExisting {
			continue
		}
		current := mapping.Content[index+1]
		if field.Fields != nil && current.Kind == yaml.MappingNode {
			current.Style = 0
			Patch(current, field.Fields)
			continue
		}
		mapping.Content[index+1] = withComments(field.node(), current)
	}
	if len(mapping.Content) > 0 {
		mapping.Style = 0
	}
}

func keyIndex(mapping *yaml.Node, key string) int {
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return index
		}
	}
	return -1
}

func removePair(mapping *yaml.Node, index int) {
	if index >= 0 {
		mapping.Content = append(mapping.Content[:index], mapping.Content[index+2:]...)
	}
}

func insertPair(mapping *yaml.Node, index int, field Field) {
	key := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: field.Key}
	pair := []*yaml.Node{key, field.node()}
	mapping.Content = append(mapping.Content[:index], append(pair, mapping.Content[index:]...)...)
}

func withComments(replacement, current *yaml.Node) *yaml.Node {
	replacement.HeadComment = current.HeadComment
	replacement.LineComment = current.LineComment
	replacement.FootComment = current.FootComment
	return replacement
}
