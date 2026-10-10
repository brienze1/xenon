package dto

import "go.yaml.in/yaml/v3"

type YamlHttp struct {
	Body    *YamlBody  `yaml:"body"`
	Method  string     `yaml:"method"`
	Url     string     `yaml:"url"`
	Params  []YamlPair `yaml:"params"`
	Headers []YamlPair `yaml:"headers"`
	Auth    yaml.Node  `yaml:"auth"`
}
