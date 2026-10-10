package dto

import "go.yaml.in/yaml/v3"

type YamlFormRow struct {
	Name        string    `yaml:"name"`
	Type        string    `yaml:"type"`
	ContentType string    `yaml:"contentType"`
	Value       yaml.Node `yaml:"value"`
	Disabled    bool      `yaml:"disabled"`
}
