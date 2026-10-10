package dto

type YamlCollection struct {
	Docs           string              `yaml:"docs"`
	OpenCollection string              `yaml:"opencollection"`
	Info           YamlInfo            `yaml:"info"`
	Request        YamlRequestSettings `yaml:"request"`
}
