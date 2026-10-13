package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiField struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
	Binary      bool   `json:"binary"`
}

func NewApiField(field entity.ApiField) ApiField {
	return ApiField{Name: field.Name, Description: field.Description, Example: field.Example, Binary: field.Binary}
}
