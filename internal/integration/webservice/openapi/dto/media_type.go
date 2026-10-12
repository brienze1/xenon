package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
)

type MediaType struct {
	Name     string   `yaml:"-"`
	Schema   *Schema  `yaml:"schema,omitempty"`
	Example  *Value   `yaml:"example,omitempty"`
	Examples Examples `yaml:"examples,omitempty"`
}

func NewBodyMediaType(body entity.ApiBody) MediaType {
	switch {
	case body.IsForm():
		return MediaType{Name: body.MediaType, Schema: NewObjectSchema(body.Fields)}
	case body.Binary:
		return MediaType{Name: body.MediaType, Schema: NewBinarySchema()}
	default:
		return MediaType{Name: body.MediaType, Example: NewValue(exampleNode(body.MediaType, body.Example))}
	}
}

func (m *MediaType) body() (entity.ApiBody, error) {
	body := entity.ApiBody{MediaType: m.Name, Binary: m.Schema.IsBinary()}
	if m.Schema != nil {
		for index := range m.Schema.Properties {
			body.Fields = append(body.Fields, m.Schema.Properties[index].ToEntity())
		}
	}
	if body.IsForm() {
		body.Binary = false
		return body, nil
	}
	example, err := exampleText(m.Example.Node())
	body.Example = example
	return body, err
}

func (m *MediaType) examples(description string) ([]entity.ApiExample, error) {
	examples := make([]entity.ApiExample, 0, len(m.Examples)+1)
	if m.Example != nil {
		single := Example{Name: description, Value: m.Example}
		example, err := single.ToEntity(m.Name)
		if err != nil {
			return nil, err
		}
		examples = append(examples, example)
	}
	for index := range m.Examples {
		example, err := m.Examples[index].ToEntity(m.Name)
		if err != nil {
			return nil, err
		}
		examples = append(examples, example)
	}
	return examples, nil
}

func (m MediaType) MarshalYAML() (any, error) {
	return mappingOfFields(
		when(m.Schema != nil, "schema", m.Schema),
		when(m.Example != nil, "example", m.Example),
		when(len(m.Examples) > 0, "examples", m.Examples),
	)
}
