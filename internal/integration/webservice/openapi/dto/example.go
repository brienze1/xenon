package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
)

type Example struct {
	Value       *Value `yaml:"value,omitempty"`
	Name        string `yaml:"-"`
	XenonId     string `yaml:"x-xenon-id,omitempty"`
	Summary     string `yaml:"summary,omitempty"`
	Description string `yaml:"description,omitempty"`
	XenonBinary bool   `yaml:"x-xenon-binary,omitempty"`
}

func NewExample(example entity.ApiExample) Example {
	if example.Binary {
		return Example{Name: example.Name, XenonId: example.ID, XenonBinary: true}
	}
	return Example{Name: example.Name, XenonId: example.ID, Value: NewValue(exampleNode(example.MediaType, example.Value))}
}

func (e *Example) ToEntity(mediaType string) (entity.ApiExample, error) {
	value, err := exampleText(e.Value.Node())
	return entity.ApiExample{ID: e.XenonId, Name: e.Name, MediaType: mediaType, Value: value, Binary: e.XenonBinary}, err
}

func (e Example) MarshalYAML() (any, error) {
	return mappingOfFields(
		text("x-xenon-id", e.XenonId),
		text("summary", e.Summary),
		text("description", e.Description),
		when(e.XenonBinary, "x-xenon-binary", e.XenonBinary),
		when(e.Value != nil, "value", e.Value),
	)
}
