package dto

import "github.com/brienze1/xenon/internal/domain/entity"

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
