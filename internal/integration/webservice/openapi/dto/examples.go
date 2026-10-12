package dto

import "go.yaml.in/yaml/v3"

type Examples []Example

func (e Examples) MarshalYAML() (any, error) {
	return mappingOf(e, func(example Example) string { return example.Name })
}

func (e *Examples) UnmarshalYAML(node *yaml.Node) error {
	entries, err := entriesOf(node, func(example *Example, name string) { example.Name = name })
	*e = entries
	return err
}
