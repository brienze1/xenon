package dto

import "go.yaml.in/yaml/v3"

type YamlBody struct {
	Type      string    `yaml:"type"`
	Query     string    `yaml:"query"`
	Variables string    `yaml:"variables"`
	Data      yaml.Node `yaml:"data"`
}
