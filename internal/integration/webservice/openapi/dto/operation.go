package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type Operation struct {
	Tags        []string               `yaml:"tags,omitempty"`
	Summary     string                 `yaml:"summary,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	XenonId     string                 `yaml:"x-xenon-id,omitempty"`
	Parameters  []Parameter            `yaml:"parameters,omitempty"`
	RequestBody *RequestBody           `yaml:"requestBody,omitempty"`
	Responses   Responses              `yaml:"responses,omitempty"`
	Security    *[]map[string][]string `yaml:"security,omitempty"`
	Servers     []Server               `yaml:"servers,omitempty"`
	Variants    []Operation            `yaml:"x-xenon-variants,omitempty"`
}

func NewOperation(operation entity.ApiOperation) Operation {
	security := make([]map[string][]string, 0, len(operation.Security))
	for _, name := range operation.Security {
		security = append(security, map[string][]string{name: {}})
	}
	return Operation{
		Tags:        operation.Tags,
		Summary:     operation.Summary,
		Description: operation.Description,
		XenonId:     operation.RequestID,
		Parameters:  NewParameters(operation),
		RequestBody: NewRequestBody(operation.Body),
		Responses:   NewResponses(operation.Responses),
		Security:    &security,
		Servers:     NewServers(operation.Server),
	}
}

// ToEntity reads the operation and its variants, which share its method and path.
func (o *Operation) ToEntity(method, path string, shared []Parameter) ([]entity.ApiOperation, error) {
	operation, err := o.operation(method, path, shared)
	if err != nil {
		return nil, err
	}
	operations := []entity.ApiOperation{operation}
	for index := range o.Variants {
		variants, err := o.Variants[index].ToEntity(method, path, shared)
		if err != nil {
			return nil, err
		}
		for _, variant := range variants {
			variant.Variant = true
			operations = append(operations, variant)
		}
	}
	return operations, nil
}

func (o *Operation) operation(method, path string, shared []Parameter) (entity.ApiOperation, error) {
	body, err := o.RequestBody.ToEntity()
	if err != nil {
		return entity.ApiOperation{}, err
	}
	responses, err := o.Responses.ToEntity()
	if err != nil {
		return entity.ApiOperation{}, err
	}
	operation := entity.ApiOperation{
		RequestID:   strings.TrimSpace(o.XenonId),
		Method:      strings.ToUpper(method),
		Path:        path,
		Server:      firstServer(o.Servers),
		Summary:     o.Summary,
		Description: o.Description,
		Body:        body,
		Tags:        append([]string{}, o.Tags...),
		PathParams:  []entity.ApiParameter{},
		QueryParams: []entity.ApiParameter{},
		Headers:     []entity.ApiParameter{},
		Responses:   responses,
		Security:    o.securityNames(),
	}
	for _, parameter := range o.parameters(shared) {
		switch {
		case parameter.isIn(inPath):
			operation.PathParams = append(operation.PathParams, parameter.ToEntity())
		case parameter.isIn(inQuery):
			operation.QueryParams = append(operation.QueryParams, parameter.ToEntity())
		case parameter.isIn(inHeader):
			operation.Headers = append(operation.Headers, parameter.ToEntity())
		}
	}
	return operation, nil
}

func (o *Operation) parameters(shared []Parameter) []Parameter {
	parameters := make([]Parameter, 0, len(shared)+len(o.Parameters))
	for _, parameter := range shared {
		overridden := false
		for _, own := range o.Parameters {
			overridden = overridden || own.sameAs(parameter)
		}
		if !overridden {
			parameters = append(parameters, parameter)
		}
	}
	return append(parameters, o.Parameters...)
}

func (o *Operation) securityNames() []string {
	names := make([]string, 0)
	if o.Security == nil {
		return names
	}
	for _, requirement := range *o.Security {
		for name := range requirement {
			names = append(names, name)
		}
	}
	return names
}

func (o Operation) MarshalYAML() (any, error) {
	return mappingOfFields(
		when(len(o.Tags) > 0, "tags", o.Tags),
		text("summary", o.Summary),
		text("description", o.Description),
		text("x-xenon-id", o.XenonId),
		when(len(o.Parameters) > 0, "parameters", o.Parameters),
		when(o.RequestBody != nil, "requestBody", o.RequestBody),
		when(len(o.Responses) > 0, "responses", o.Responses),
		when(o.Security != nil, "security", o.Security),
		when(len(o.Servers) > 0, "servers", o.Servers),
		when(len(o.Variants) > 0, "x-xenon-variants", o.Variants),
	)
}
