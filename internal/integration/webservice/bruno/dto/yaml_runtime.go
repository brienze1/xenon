package dto

type YamlRuntime struct {
	Variables  []YamlPair      `yaml:"variables"`
	Assertions []YamlAssertion `yaml:"assertions"`
	Scripts    []YamlScript    `yaml:"scripts"`
}
