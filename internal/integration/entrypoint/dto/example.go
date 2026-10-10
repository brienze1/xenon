package dto

import (
	"encoding/base64"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
)

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

type ExampleResponse struct {
	Body         *string    `json:"body,omitempty"`
	Status       string     `json:"status"`
	ContentType  string     `json:"content_type"`
	BodyEncoding string     `json:"body_encoding"`
	BodyUrl      string     `json:"body_url"`
	Headers      []KeyValue `json:"headers"`
	Code         int        `json:"code"`
	SizeBytes    int        `json:"size_bytes"`
	IsBinary     bool       `json:"is_binary"`
}

func NewExampleResponse(details entity.ExampleDetails) ExampleResponse {
	response := details.Example.Response
	result := ExampleResponse{
		Code:         response.Code,
		Status:       response.Status,
		ContentType:  response.ContentType,
		Headers:      NewKeyValues(response.Headers),
		BodyEncoding: response.BodyEncoding.String(),
		SizeBytes:    len(response.Body),
		IsBinary:     response.IsBinary(),
	}
	if details.ServesBodyByRoute() {
		result.BodyUrl = ResponseBodyRoute + "/" + details.Example.BodyKey()
		return result
	}
	body := string(response.Body)
	result.Body = &body
	return result
}

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

type ExampleResponseEdit struct {
	Body         *string    `json:"body"`
	Status       string     `json:"status"`
	ContentType  string     `json:"content_type"`
	BodyEncoding string     `json:"body_encoding" validate:"omitempty,oneof=text base64"`
	Headers      []KeyValue `json:"headers"`
	Code         int        `json:"code"`
}

func (e *ExampleResponseEdit) ToEntity() entity.ExampleResponse {
	response := entity.NewExampleResponse()
	response.Code = e.Code
	response.Status = e.Status
	response.ContentType = e.ContentType
	response.Headers = keyValuesToEntity(e.Headers)
	if e.Body == nil {
		return response
	}
	if bodyencoding.GetBodyEncoding(e.BodyEncoding) != bodyencoding.Base64 {
		return response.WithBody([]byte(*e.Body))
	}
	if decoded, err := base64.StdEncoding.DecodeString(*e.Body); err == nil {
		return response.WithBinaryBody(decoded)
	}
	return response.WithBody([]byte(*e.Body))
}

type ExampleEdit struct {
	Id       string              `json:"id" validate:"required,not_blank"`
	Name     string              `json:"name"`
	Request  ExampleRequest      `json:"request"`
	Response ExampleResponseEdit `json:"response"`
}

func (e *ExampleEdit) ToEntity() entity.SavedExample {
	return entity.SavedExample{ID: e.Id, Name: e.Name, Request: e.Request.ToEntity(), Response: e.Response.ToEntity()}
}
