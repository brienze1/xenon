package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SendRequest struct {
	LocalId string  `json:"local_id" validate:"required"`
	TabId   string  `json:"tab_id" validate:"required"`
	Request Request `json:"request"`
}

func (s SendRequest) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "tab_id": s.TabId, "request": s.Request.LogValue()}
}

func (s *SendRequest) ToEntity() entity.Request {
	return s.Request.ToEntity()
}
