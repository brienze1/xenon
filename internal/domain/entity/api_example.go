package entity

import (
	"encoding/json"
	"net/http"
	"strings"
)

const defaultExampleName = "Example"

type ApiExample struct {
	ID        string
	Name      string
	MediaType string
	Value     string
	Binary    bool
}

func NewApiExample(example SavedExample) ApiExample {
	documented := ApiExample{ID: example.ID, Name: example.Name, MediaType: exampleMediaType(example.Response)}
	if example.Response.IsBinary() {
		documented.Binary = true
		return documented
	}
	documented.Value = string(example.Response.Body)
	return documented
}

// AppliedTo writes the documented example over a saved one; a binary body stays as it is.
func (a ApiExample) AppliedTo(example SavedExample, code int, baseline ApiExample) SavedExample {
	example.Name = a.nameOr(example.Name)
	example.Response.Code = code
	example.Response.Status = http.StatusText(code)
	if a.MediaType != baseline.MediaType {
		example.Response.ContentType = a.MediaType
	}
	if !a.Binary && (a.Value != baseline.Value || baseline.Binary) {
		example.Response = example.Response.WithBody([]byte(a.Value))
	}
	return example
}

func (a ApiExample) NewSavedExample(id string, code int, request Request) SavedExample {
	example := NewExampleOf(id, a.nameOr(defaultExampleName), request)
	example.Response.Code = code
	example.Response.Status = http.StatusText(code)
	example.Response.ContentType = a.MediaType
	if !a.Binary {
		example.Response = example.Response.WithBody([]byte(a.Value))
	}
	return example
}

func (a ApiExample) nameOr(fallback string) string {
	if name := NormalizeItemName(a.Name); name != "" {
		return name
	}
	return fallback
}

func exampleMediaType(response ExampleResponse) string {
	mediaType, _, _ := strings.Cut(response.ContentType, ";")
	if mediaType = strings.TrimSpace(mediaType); mediaType != "" {
		return mediaType
	}
	if response.IsBinary() {
		return MediaTypeOctetStream
	}
	if json.Valid(response.Body) {
		return MediaTypeJson
	}
	return MediaTypeText
}
