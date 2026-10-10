package dto

import "go.yaml.in/yaml/v3"

type YamlRequest struct {
	Http     YamlHttp    `yaml:"http"`
	Docs     string      `yaml:"docs"`
	Info     YamlInfo    `yaml:"info"`
	Runtime  YamlRuntime `yaml:"runtime"`
	Settings yaml.Node   `yaml:"settings"`
}
