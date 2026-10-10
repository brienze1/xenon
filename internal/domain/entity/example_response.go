package entity

import (
	"slices"
	"unicode/utf8"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/bodyencoding"
)

const (
	minResponseCode = 100
	maxResponseCode = 599
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

func NewExampleResponseOf(response Response, body []byte) ExampleResponse {
	example := ExampleResponse{Code: response.StatusCode, Status: response.StatusText, ContentType: response.ContentType, Headers: slices.Clone(response.Headers)}
	if response.Binary {
		return example.WithBinaryBody(body)
	}
	return example.WithBody(body)
}

func (e ExampleResponse) WithBody(body []byte) ExampleResponse {
	e.Body = body
	e.BodyEncoding = bodyencoding.Text
	if !utf8.Valid(body) {
		e.BodyEncoding = bodyencoding.Base64
	}
	return e
}

func (e ExampleResponse) WithBinaryBody(body []byte) ExampleResponse {
	e.Body = body
	e.BodyEncoding = bodyencoding.Base64
	return e
}

func (e ExampleResponse) WithBodyOf(stored ExampleResponse) ExampleResponse {
	e.Body = stored.Body
	e.BodyEncoding = stored.BodyEncoding
	return e
}

func (e ExampleResponse) IsBinary() bool {
	return e.BodyEncoding == bodyencoding.Base64
}

func (e ExampleResponse) HasValidCode() bool {
	return e.Code >= minResponseCode && e.Code <= maxResponseCode
}

func (e ExampleResponse) HasInlineBody() bool {
	return !e.IsBinary() && len(e.Body) <= InlineBodyLimit
}

func (e ExampleResponse) AsResponse() Response {
	response := NewResponse(e.Code, e.Status, "", e.Headers, e.Body, int64(len(e.Body)), Timings{})
	response.ContentType = e.ContentType
	response.Binary = e.IsBinary()
	return response
}
