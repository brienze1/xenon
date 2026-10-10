package entity

import "strings"

type SavedExample struct {
	Postman  SourceFields
	Request  ExampleRequest
	ID       string
	Name     string
	Response ExampleResponse
}

func NewSavedExample(id, name string) SavedExample {
	return SavedExample{ID: id, Name: strings.TrimSpace(name), Request: NewExampleRequest(), Response: NewExampleResponse()}
}

func (s SavedExample) LogValue() any {
	return map[string]any{"id": s.ID, "name": s.Name, "code": s.Response.Code}
}
