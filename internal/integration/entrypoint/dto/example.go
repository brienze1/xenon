package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Example struct {
	Id        string          `json:"id"`
	RequestId string          `json:"request_id"`
	Name      string          `json:"name"`
	Revision  string          `json:"revision"`
	Request   ExampleRequest  `json:"request"`
	Response  ExampleResponse `json:"response"`
}

func NewExample(details entity.ExampleDetails) Example {
	return Example{
		Id:        details.Example.ID,
		RequestId: details.RequestID,
		Name:      details.Example.Name,
		Revision:  details.Revision,
		Request:   NewExampleRequest(details.Example.Request),
		Response:  NewExampleResponse(details),
	}
}

func (e Example) LogValue() any {
	return map[string]any{"id": e.Id, "request_id": e.RequestId, "code": e.Response.Code}
}
