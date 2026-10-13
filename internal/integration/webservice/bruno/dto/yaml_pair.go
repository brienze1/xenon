package dto

import "go.yaml.in/yaml/v3"

type YamlPair struct {
	Enabled  *bool  `yaml:"enabled"`
	Name     string `yaml:"name"`
	Value    string `yaml:"value"`
	Type     string `yaml:"type"`
	Scope    string `yaml:"scope"`
	Line     int    `yaml:"-"`
	Disabled bool   `yaml:"disabled"`
	Secret   bool   `yaml:"secret"`
}

func (y *YamlPair) UnmarshalYAML(node *yaml.Node) error {
	type plain YamlPair
	if err := node.Decode((*plain)(y)); err != nil {
		return err
	}
	y.Line = node.Line
	return nil
}

func (y YamlPair) IsEnabled() bool {
	return !y.Disabled && (y.Enabled == nil || *y.Enabled)
}

func (y YamlPair) ToPair() Pair {
	return NewPair(y.Name, y.Value, y.Line, y.IsEnabled())
}
