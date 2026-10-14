package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
)

func NewStoredExampleResponseEdit(response entity.ExampleResponse) ExampleResponseEdit {
	return ExampleResponseEdit{Status: response.Status, ContentType: response.ContentType, Headers: NewKeyValues(response.Headers), Code: response.Code}
}

func NewTextExampleResponseEdit(response ExampleResponseEdit, body string) ExampleResponseEdit {
	return ExampleResponseEdit{
		Body:         &body,
		Status:       response.Status,
		ContentType:  response.ContentType,
		BodyEncoding: bodyencoding.Text.String(),
		Headers:      response.Headers,
		Code:         response.Code,
	}
}
