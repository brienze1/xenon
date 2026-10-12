package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const (
	typeString   = "string"
	typeObject   = "object"
	formatBinary = "binary"
)

type Schema struct {
	Type        string     `yaml:"type,omitempty"`
	Format      string     `yaml:"format,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Example     string     `yaml:"example,omitempty"`
	Properties  Properties `yaml:"properties,omitempty"`
}

func NewStringSchema() *Schema {
	return &Schema{Type: typeString}
}

func NewBinarySchema() *Schema {
	return &Schema{Type: typeString, Format: formatBinary}
}

func NewFieldSchema(field entity.ApiField) Schema {
	if field.Binary {
		return Schema{Type: typeString, Format: formatBinary, Description: field.Description}
	}
	return Schema{Type: typeString, Description: field.Description, Example: field.Example}
}

func NewObjectSchema(fields []entity.ApiField) *Schema {
	properties := make(Properties, 0, len(fields))
	for _, field := range fields {
		properties = append(properties, NewProperty(field))
	}
	return &Schema{Type: typeObject, Properties: properties}
}

func (s *Schema) IsBinary() bool {
	return s != nil && s.Format == formatBinary
}

func (s Schema) MarshalYAML() (any, error) {
	return mappingOfFields(
		text("type", s.Type),
		text("format", s.Format),
		text("description", s.Description),
		when(len(s.Properties) > 0, "properties", s.Properties),
		text("example", s.Example),
	)
}
