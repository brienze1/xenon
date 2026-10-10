package dto

import (
	"encoding/json"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

var (
	responseKeys     = []string{"name", "originalRequest", "code", "status", "header", "body"}
	previewMimeTypes = map[string]string{
		"json":       "application/json",
		"xml":        "application/xml",
		"html":       "text/html",
		"text":       "text/plain",
		"javascript": "application/javascript",
	}
)

type Response struct {
	Raw             map[string]any `json:"-"`
	OriginalRequest *Request       `json:"originalRequest"`
	Body            any            `json:"body"`
	Name            string         `json:"name"`
	Status          string         `json:"status"`
	PreviewLanguage string         `json:"_postman_previewlanguage"`
	Header          HeaderList     `json:"header"`
	Code            int            `json:"code"`
}

func (r *Response) UnmarshalJSON(data []byte) error {
	type plain Response
	if err := json.Unmarshal(data, (*plain)(r)); err != nil {
		return err
	}
	r.Raw = decodeRaw(data)
	return nil
}

func (r *Response) ToEntity(id string) entity.SavedExample {
	example := entity.NewSavedExample(id, r.Name)
	if r.OriginalRequest != nil {
		example.Request = r.OriginalRequest.ToExampleRequest()
	}
	example.Response.Code = r.Code
	example.Response.Status = r.Status
	example.Response.Headers = r.Header.ToEntity()
	example.Response.ContentType = r.contentType(example.Response.Headers)
	example.Response = example.Response.WithBody([]byte(text(r.Body)))
	example.Postman = r.Passthrough()
	return example
}

func (r *Response) Passthrough() map[string]any {
	fields := r.Header.Passthrough(without(r.Raw, responseKeys...))
	if r.OriginalRequest != nil {
		fields = withEntry(fields, "originalRequest", r.OriginalRequest.Passthrough())
	}
	return fields
}

func (r *Response) contentType(headers []entity.KeyValue) string {
	for _, header := range headers {
		if strings.EqualFold(header.Key, "Content-Type") {
			return header.Value
		}
	}
	return previewMimeTypes[r.PreviewLanguage]
}
