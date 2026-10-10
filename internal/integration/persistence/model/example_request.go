package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type ExampleRequest struct {
	Auth          Auth           `yaml:"auth"`
	Body          Body           `yaml:"body"`
	Method        string         `yaml:"method"`
	Url           string         `yaml:"url"`
	Params        []KeyValue     `yaml:"params"`
	PathVariables []PathVariable `yaml:"pathVariables"`
	Headers       []KeyValue     `yaml:"headers"`
}

func NewExampleRequest(request entity.ExampleRequest) ExampleRequest {
	model := ExampleRequest{
		Method:        request.Method,
		Url:           request.URL,
		Params:        newKeyValues(request.Params),
		PathVariables: make([]PathVariable, 0, len(request.PathVariables)),
		Headers:       newKeyValues(request.Headers),
		Body:          NewBody(request.Body),
		Auth:          NewAuth(request.Auth),
	}
	for _, variable := range request.PathVariables {
		model.PathVariables = append(model.PathVariables, NewPathVariable(variable))
	}
	return model
}

func (e *ExampleRequest) ToEntity() entity.ExampleRequest {
	request := entity.NewExampleRequest()
	if e.Method != "" {
		request.Method = e.Method
	}
	request.URL = e.Url
	request.Params = keyValuesToEntity(e.Params)
	for index := range e.PathVariables {
		request.PathVariables = append(request.PathVariables, e.PathVariables[index].ToEntity())
	}
	request.Headers = keyValuesToEntity(e.Headers)
	request.Body = e.Body.ToEntity()
	request.Auth = e.Auth.ToEntity()
	return request
}

func (e *ExampleRequest) Fields() []yamlnode.Field {
	return []yamlnode.Field{
		yamlnode.String("method", e.Method),
		yamlnode.String("url", e.Url),
		yamlnode.List("params", keyValueNodes(e.Params)),
		yamlnode.List("pathVariables", pathVariableNodes(e.PathVariables)),
		yamlnode.List("headers", keyValueNodes(e.Headers)),
		e.Body.Field(),
		e.Auth.Field(),
	}
}
