package entity

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums/httpmethod"
)

type ExampleRequest struct {
	Auth          Auth
	Body          Body
	Method        string
	URL           string
	Params        []KeyValue
	PathVariables []PathVariable
	Headers       []KeyValue
}

func NewExampleRequest() ExampleRequest {
	return ExampleRequest{
		Method:        httpmethod.Get.String(),
		Params:        []KeyValue{},
		PathVariables: []PathVariable{},
		Headers:       []KeyValue{},
		Auth:          DefaultAuth(),
		Body:          DefaultBody(),
	}
}

func NewExampleRequestOf(request Request) ExampleRequest {
	return ExampleRequest{
		Method:        request.MethodOrDefault(),
		URL:           request.URL,
		Params:        slices.Clone(request.Params),
		PathVariables: slices.Clone(request.PathVariables),
		Headers:       slices.Clone(request.Headers),
		Auth:          request.Auth,
		Body:          request.Body,
	}
}
