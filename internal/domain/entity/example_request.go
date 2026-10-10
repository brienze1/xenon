package entity

import "github.com/brienze1/xenon/internal/domain/enums/httpmethod"

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
