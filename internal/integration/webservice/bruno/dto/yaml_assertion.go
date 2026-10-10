package dto

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type YamlAssertion struct {
	Enabled    *bool  `yaml:"enabled"`
	Expression string `yaml:"expression"`
	Operator   string `yaml:"operator"`
	Value      string `yaml:"value"`
	Line       int    `yaml:"-"`
	Disabled   bool   `yaml:"disabled"`
}

func (y *YamlAssertion) UnmarshalYAML(node *yaml.Node) error {
	type plain YamlAssertion
	if err := node.Decode((*plain)(y)); err != nil {
		return err
	}
	y.Line = node.Line
	return nil
}

func (y YamlAssertion) IsEnabled() bool {
	return !y.Disabled && (y.Enabled == nil || *y.Enabled)
}

func (y YamlAssertion) ToEntity() entity.BrunoAssertion {
	return entity.BrunoAssertion{Expression: y.Expression, Operator: y.Operator, Value: y.Value, Line: y.Line, Enabled: y.IsEnabled()}
}
