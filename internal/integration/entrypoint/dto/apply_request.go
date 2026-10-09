package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApplyRequest struct {
	Request Request `json:"request" validate:"required"`
}

func (a ApplyRequest) LogValue() any {
	return a.Request.LogValue()
}

func (a *ApplyRequest) ToEntity() entity.Request {
	return a.Request.ToEntity()
}
