package model

import (
	"encoding/base64"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type ExampleResponse struct {
	Status       string     `yaml:"status"`
	ContentType  string     `yaml:"contentType"`
	BodyEncoding string     `yaml:"bodyEncoding"`
	Body         string     `yaml:"body"`
	Headers      []KeyValue `yaml:"headers"`
	Code         int        `yaml:"code"`
}

func NewExampleResponse(response entity.ExampleResponse) ExampleResponse {
	model := ExampleResponse{
		Code:         response.Code,
		Status:       response.Status,
		ContentType:  response.ContentType,
		Headers:      newKeyValues(response.Headers),
		BodyEncoding: bodyencoding.Text.String(),
		Body:         string(response.Body),
	}
	if response.IsBinary() {
		model.BodyEncoding = bodyencoding.Base64.String()
		model.Body = base64.StdEncoding.EncodeToString(response.Body)
	}
	return model
}

func (e *ExampleResponse) ToEntity() entity.ExampleResponse {
	response := entity.NewExampleResponse()
	response.Code = e.Code
	response.Status = e.Status
	response.ContentType = e.ContentType
	response.Headers = keyValuesToEntity(e.Headers)
	response.Body = []byte(e.Body)
	if bodyencoding.GetBodyEncoding(e.BodyEncoding) == bodyencoding.Base64 {
		decoded, _ := base64.StdEncoding.DecodeString(e.Body)
		response.Body = decoded
		response.BodyEncoding = bodyencoding.Base64
	}
	return response
}

func (e *ExampleResponse) Fields() []yamlnode.Field {
	return []yamlnode.Field{
		yamlnode.Int("code", e.Code),
		yamlnode.String("status", e.Status),
		yamlnode.String("contentType", e.ContentType),
		yamlnode.List("headers", keyValueNodes(e.Headers)),
		yamlnode.String("bodyEncoding", e.BodyEncoding),
		yamlnode.String("body", e.Body),
	}
}
