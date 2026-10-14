package dto

import "go.yaml.in/yaml/v3"

const yamlNullTag = "!!null"

type KeyValues []KeyValue

func (k *KeyValues) UnmarshalYAML(node *yaml.Node) error {
	switch {
	case node.Kind == yaml.SequenceNode:
		var rows []KeyValue
		if err := node.Decode(&rows); err != nil {
			return err
		}
		*k = rows
	case node.Kind == yaml.MappingNode:
		rows := make([]KeyValue, 0, len(node.Content)/2)
		for index := 0; index+1 < len(node.Content); index += 2 {
			var value any
			if err := node.Content[index+1].Decode(&value); err != nil {
				return err
			}
			rows = append(rows, NewKeyValue(node.Content[index].Value, value))
		}
		*k = rows
	case node.Kind == yaml.ScalarNode && node.Tag == yamlNullTag:
		*k = KeyValues{}
	default:
		return NewWrongTypeError(node.Line)
	}
	return nil
}

func (k KeyValues) toPostmanV21() []any {
	rows := make([]any, 0, len(k))
	for index := range k {
		rows = append(rows, k[index].toPostmanV21())
	}
	return rows
}

func (k KeyValues) toPathVariables() []any {
	rows := make([]any, 0, len(k))
	for index := range k {
		row := k[index].toPostmanV21()
		delete(row, "disabled")
		rows = append(rows, row)
	}
	return rows
}
