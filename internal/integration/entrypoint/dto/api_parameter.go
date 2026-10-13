package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Example     string `json:"example"`
	Required    bool   `json:"required"`
	Variable    bool   `json:"variable"`
}

func NewApiParameter(parameter entity.ApiParameter, in string) ApiParameter {
	return ApiParameter{Name: parameter.Name, In: in, Description: parameter.Description, Example: parameter.Example, Required: parameter.Required, Variable: parameter.Variable}
}
