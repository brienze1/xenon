package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RequestAlias struct {
	Id      string  `json:"id" validate:"required,not_blank"`
	Name    string  `json:"name"`
	Request Request `json:"request"`
}

func (r *RequestAlias) ToEntity() entity.RequestAlias {
	return entity.RequestAlias{ID: r.Id, Name: r.Name, Version: r.Request.ToEntity()}
}

func (r RequestAlias) LogValue() any {
	return map[string]any{"id": r.Id, "name": r.Name, "request": r.Request.LogValue()}
}
