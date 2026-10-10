package dto

type YamlInfo struct {
	Seq  *int     `yaml:"seq"`
	Name string   `yaml:"name"`
	Type string   `yaml:"type"`
	Tags []string `yaml:"tags"`
}
