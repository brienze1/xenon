package dto

import (
	"encoding/base64"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
)

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
