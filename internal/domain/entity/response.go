package entity

import (
	"bytes"
	"mime"
	"strings"
	"unicode/utf8"
)

const (
	InlineBodyLimit = 1 << 20
	MemoryBodyLimit = 5 << 20
)

var textualMediaTypes = []string{"json", "xml", "javascript", "x-www-form-urlencoded", "graphql", "yaml"}

type Response struct {
	Error       *SendError
	Sent        *Request
	ContentType string
	StatusText  string
	Proto       string
	Headers     []KeyValue
	Body        []byte
	Timings     Timings
	StatusCode  int
	SizeBytes   int64
	BodyLength  int64
	Binary      bool
}

func NewResponse(statusCode int, statusText, proto string, headers []KeyValue, body []byte, sizeBytes int64, timings Timings) Response {
	response := Response{
		StatusCode: statusCode,
		StatusText: statusText,
		Proto:      proto,
		Headers:    headers,
		Body:       body,
		SizeBytes:  sizeBytes,
		BodyLength: int64(len(body)),
		Timings:    timings,
	}
	response.ContentType = response.Header(HeaderContentType)
	response.Binary = isBinaryBody(response.ContentType, body)
	return response
}

func NewFailedResponse(sendError SendError, timings Timings) Response {
	return Response{Error: &sendError, Headers: []KeyValue{}, Body: []byte{}, Timings: timings}
}

func (r Response) IsOk() bool {
	return r.Error == nil
}

func (r Response) Header(name string) string {
	for _, header := range r.Headers {
		if header.HasKey(name) {
			return header.Value
		}
	}
	return ""
}

func (r Response) HasInlineBody() bool {
	return r.IsOk() && !r.Binary && r.BodyLength <= InlineBodyLimit
}

func (r Response) KeepsBodyInMemory() bool {
	return r.BodyLength <= MemoryBodyLimit
}

func (r Response) WithMaskedValues(values []string) Response {
	if r.Error != nil {
		masked := r.Error.Masked(values)
		r.Error = &masked
	}
	return r
}

func (r Response) WithSentRequest(request Request) Response {
	r.Sent = &request
	return r
}

func (r Response) SentRequestOr(fallback Request) Request {
	if r.Sent != nil {
		return *r.Sent
	}
	return fallback
}

func (r Response) WithoutBody() Response {
	r.Body = nil
	return r
}

func (r Response) LogValue() any {
	return map[string]any{"statusCode": r.StatusCode, "sizeBytes": r.SizeBytes, "ok": r.IsOk()}
}

func isBinaryBody(contentType string, body []byte) bool {
	if isTextualContentType(contentType) {
		return false
	}
	return bytes.Contains(body, []byte{0}) || !utf8.Valid(body)
}

func isTextualContentType(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}
	for _, textual := range textualMediaTypes {
		if strings.Contains(mediaType, textual) {
			return true
		}
	}
	return false
}
