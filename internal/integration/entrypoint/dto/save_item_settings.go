package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SaveItemSettings struct {
	Auth        Auth       `json:"auth"`
	LocalId     string     `json:"local_id" validate:"required,not_blank"`
	Id          string     `json:"id" validate:"required,not_blank"`
	Description string     `json:"description"`
	Headers     []KeyValue `json:"headers"`
}

func (s *SaveItemSettings) ToEntity() entity.Collection {
	return entity.Collection{ID: s.Id, Description: s.Description, Auth: s.Auth.ToEntity(), Headers: keyValuesToEntity(s.Headers)}
}

func (s SaveItemSettings) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "id": s.Id}
}
