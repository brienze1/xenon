package dto

import "go.yaml.in/yaml/v3"

type YamlRequestSettings struct {
	Headers   []YamlPair      `yaml:"headers"`
	Variables []YamlPair      `yaml:"variables"`
	Scripts   []YamlScript    `yaml:"scripts"`
	Asserts   []YamlAssertion `yaml:"assertions"`
	Auth      yaml.Node       `yaml:"auth"`
}
