package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RequestBody struct {
	Description string  `yaml:"description,omitempty"`
	Content     Content `yaml:"content"`
}

func NewRequestBody(body entity.ApiBody) *RequestBody {
	if !body.HasContent() {
		return nil
	}
	return &RequestBody{Content: Content{NewBodyMediaType(body)}}
}

func (r *RequestBody) ToEntity() (entity.ApiBody, error) {
	if r == nil || len(r.Content) == 0 {
		return entity.ApiBody{}, nil
	}
	return r.Content[0].body()
}

func (r RequestBody) MarshalYAML() (any, error) {
	return mappingOfFields(text("description", r.Description), always("content", r.Content))
}
