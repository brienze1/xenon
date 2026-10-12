package dto

type Components struct {
	SecuritySchemes SecuritySchemes `yaml:"securitySchemes,omitempty"`
}

func (c Components) MarshalYAML() (any, error) {
	return mappingOfFields(when(len(c.SecuritySchemes) > 0, "securitySchemes", c.SecuritySchemes))
}
