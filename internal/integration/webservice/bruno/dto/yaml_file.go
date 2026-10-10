package dto

type YamlFile struct {
	FilePath    string `yaml:"filePath"`
	ContentType string `yaml:"contentType"`
	Selected    bool   `yaml:"selected"`
}
