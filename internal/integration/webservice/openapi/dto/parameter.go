package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	inPath   = "path"
	inQuery  = "query"
	inHeader = "header"
)

type Parameter struct {
	Schema        *Schema `yaml:"schema,omitempty"`
	Name          string  `yaml:"name"`
	In            string  `yaml:"in"`
	Description   string  `yaml:"description,omitempty"`
	Example       string  `yaml:"example,omitempty"`
	Required      bool    `yaml:"required,omitempty"`
	XenonVariable bool    `yaml:"x-xenon-variable,omitempty"`
}

func NewParameter(parameter entity.ApiParameter, in string) Parameter {
	return Parameter{
		Name:          parameter.Name,
		In:            in,
		Description:   parameter.Description,
		Required:      parameter.Required,
		Schema:        NewStringSchema(),
		Example:       parameter.Example,
		XenonVariable: parameter.Variable,
	}
}

func NewParameters(operation entity.ApiOperation) []Parameter {
	parameters := make([]Parameter, 0, len(operation.PathParams)+len(operation.QueryParams)+len(operation.Headers))
	for _, parameter := range operation.PathParams {
		parameters = append(parameters, NewParameter(parameter, inPath))
	}
	for _, parameter := range operation.QueryParams {
		parameters = append(parameters, NewParameter(parameter, inQuery))
	}
	for _, parameter := range operation.Headers {
		parameters = append(parameters, NewParameter(parameter, inHeader))
	}
	return parameters
}

func (p *Parameter) ToEntity() entity.ApiParameter {
	return entity.ApiParameter{Name: p.Name, Description: p.Description, Example: p.Example, Required: p.Required, Variable: p.XenonVariable}
}

func (p *Parameter) isIn(in string) bool {
	return strings.EqualFold(p.In, in)
}

func (p *Parameter) sameAs(other Parameter) bool {
	return p.Name == other.Name && p.isIn(other.In)
}

func (p Parameter) MarshalYAML() (any, error) {
	return mappingOfFields(
		always("name", p.Name),
		always("in", p.In),
		text("description", p.Description),
		when(p.Required, "required", p.Required),
		when(p.Schema != nil, "schema", p.Schema),
		text("example", p.Example),
		when(p.XenonVariable, "x-xenon-variable", p.XenonVariable),
	)
}
