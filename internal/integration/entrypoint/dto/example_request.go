package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ExampleRequest struct {
	Auth          Auth           `json:"auth"`
	Method        string         `json:"method"`
	Url           string         `json:"url"`
	Body          Body           `json:"body"`
	Params        []KeyValue     `json:"params"`
	PathVariables []PathVariable `json:"path_variables"`
	Headers       []KeyValue     `json:"headers"`
}

func NewExampleRequest(request entity.ExampleRequest) ExampleRequest {
	result := ExampleRequest{
		Method:        request.Method,
		Url:           request.URL,
		Params:        NewKeyValues(request.Params),
		PathVariables: make([]PathVariable, 0, len(request.PathVariables)),
		Headers:       NewKeyValues(request.Headers),
		Body:          NewBody(request.Body),
		Auth:          NewAuth(request.Auth),
	}
	for _, variable := range request.PathVariables {
		result.PathVariables = append(result.PathVariables, NewPathVariable(variable))
	}
	return result
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
