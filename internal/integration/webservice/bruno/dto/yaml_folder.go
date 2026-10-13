package dto

type YamlFolder struct {
	Docs    string              `yaml:"docs"`
	Info    YamlInfo            `yaml:"info"`
	Request YamlRequestSettings `yaml:"request"`
}
