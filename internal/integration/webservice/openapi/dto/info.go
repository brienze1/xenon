package dto

type Info struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version"`
}

func (i Info) MarshalYAML() (any, error) {
	return mappingOfFields(always("title", i.Title), text("description", i.Description), always("version", i.Version))
}
