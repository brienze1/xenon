package dto

import "go.yaml.in/yaml/v3"

type SecuritySchemes []SecurityScheme

func (s SecuritySchemes) MarshalYAML() (any, error) {
	return mappingOf(s, func(scheme SecurityScheme) string { return scheme.Name })
}

func (s *SecuritySchemes) UnmarshalYAML(node *yaml.Node) error {
	entries, err := entriesOf(node, func(scheme *SecurityScheme, name string) { scheme.Name = name })
	*s = entries
	return err
}
