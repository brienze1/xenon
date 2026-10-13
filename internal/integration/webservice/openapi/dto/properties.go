package dto

import "go.yaml.in/yaml/v3"

type Properties []Property

func (p Properties) MarshalYAML() (any, error) {
	return mappingOf(p, func(property Property) string { return property.Name })
}

func (p *Properties) UnmarshalYAML(node *yaml.Node) error {
	entries, err := entriesOf(node, func(property *Property, name string) { property.Name = name })
	*p = entries
	return err
}
