package entity

import "strings"

const exampleBodyKeyPrefix = "example:"

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

func NewExampleOf(id, name string, request Request) SavedExample {
	example := NewSavedExample(id, name)
	example.Request = NewExampleRequestOf(request)
	return example
}

func NewExampleFromExchange(id, name string, sent Request, response Response, body []byte) SavedExample {
	example := NewExampleOf(id, name, sent)
	example.Response = NewExampleResponseOf(response, body)
	return example
}

func ExampleBodyKey(id string) string {
	return exampleBodyKeyPrefix + id
}

func (s SavedExample) HasValidName() bool {
	return isValidItemName(s.Name)
}

func (s SavedExample) Renamed(name string) SavedExample {
	s.Name = strings.TrimSpace(name)
	return s
}

func (s SavedExample) Duplicated(id string) SavedExample {
	s.ID = id
	s.Name = NormalizeItemName(s.Name + copySuffix)
	return s
}

func (s SavedExample) Edited(edit SavedExample, keepBody bool) SavedExample {
	edit.Postman = s.Postman
	edit.Name = strings.TrimSpace(edit.Name)
	if keepBody {
		edit.Response = edit.Response.WithBodyOf(s.Response)
	}
	return edit
}

func (s SavedExample) BodyKey() string {
	return ExampleBodyKey(s.ID)
}

func (s SavedExample) TreeExample() TreeExample {
	return TreeExample{ID: s.ID, Name: s.Name, Code: s.Response.Code}
}

func (s SavedExample) LogValue() any {
	return map[string]any{"id": s.ID, "name": s.Name, "code": s.Response.Code}
}
