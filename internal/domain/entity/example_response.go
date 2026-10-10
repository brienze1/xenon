package entity

import (
	"unicode/utf8"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
)

type ExampleResponse struct {
	BodyEncoding enums.BodyEncoding
	Status       string
	ContentType  string
	Headers      []KeyValue
	Body         []byte
	Code         int
}

func NewExampleResponse() ExampleResponse {
	return ExampleResponse{Code: 200, Status: "OK", Headers: []KeyValue{}, Body: []byte{}, BodyEncoding: bodyencoding.Text}
}

func (e ExampleResponse) WithBody(body []byte) ExampleResponse {
	e.Body = body
	e.BodyEncoding = bodyencoding.Text
	if !utf8.Valid(body) {
		e.BodyEncoding = bodyencoding.Base64
	}
	return e
}

func (e ExampleResponse) IsBinary() bool {
	return e.BodyEncoding == bodyencoding.Base64
}
