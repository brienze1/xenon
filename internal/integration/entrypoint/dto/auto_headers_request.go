package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AutoHeadersRequest struct {
	LocalId string  `json:"local_id" validate:"required"`
	Request Request `json:"request"`
}

func (a AutoHeadersRequest) LogValue() any {
	return map[string]any{"local_id": a.LocalId, "request": a.Request.LogValue()}
}

func (a *AutoHeadersRequest) ToEntity() entity.Request {
	return a.Request.ToEntity()
}
