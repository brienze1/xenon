package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type DeleteCookie struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Domain  string `json:"domain" validate:"required,not_blank"`
	Path    string `json:"path"`
	Name    string `json:"name" validate:"required"`
}

func (d DeleteCookie) ToEntity() entity.Cookie {
	return entity.Cookie{Domain: d.Domain, Path: d.Path, Name: d.Name}
}
