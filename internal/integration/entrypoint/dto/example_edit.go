package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ExampleEdit struct {
	Id       string              `json:"id" validate:"required,not_blank"`
	Name     string              `json:"name"`
	Request  ExampleRequest      `json:"request"`
	Response ExampleResponseEdit `json:"response"`
}

func (e *ExampleEdit) ToEntity() entity.SavedExample {
	return entity.SavedExample{ID: e.Id, Name: e.Name, Request: e.Request.ToEntity(), Response: e.Response.ToEntity()}
}
