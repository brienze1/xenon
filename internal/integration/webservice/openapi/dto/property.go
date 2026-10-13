package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Property struct {
	Name   string `yaml:"-"`
	Schema `yaml:",inline"`
}

func NewProperty(field entity.ApiField) Property {
	return Property{Name: field.Name, Schema: NewFieldSchema(field)}
}

func (p *Property) ToEntity() entity.ApiField {
	return entity.ApiField{Name: p.Name, Description: p.Description, Example: p.Example, Binary: p.IsBinary()}
}
