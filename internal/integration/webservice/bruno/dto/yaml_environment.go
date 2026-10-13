package dto

type YamlEnvironment struct {
	Name      string     `yaml:"name"`
	Color     string     `yaml:"color"`
	Variables []YamlPair `yaml:"variables"`
}
