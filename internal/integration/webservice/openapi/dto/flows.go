package dto

import "go.yaml.in/yaml/v3"

type Flows []Flow

func (f Flows) MarshalYAML() (any, error) {
	return mappingOf(f, func(flow Flow) string { return flow.Name })
}

func (f *Flows) UnmarshalYAML(node *yaml.Node) error {
	entries, err := entriesOf(node, func(flow *Flow, name string) { flow.Name = name })
	*f = entries
	return err
}
