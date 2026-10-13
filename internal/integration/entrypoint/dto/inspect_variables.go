package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type InspectVariables struct {
	LocalId string  `json:"local_id" validate:"required,not_blank"`
	Request Request `json:"request"`
}

func (i *InspectVariables) ToEntity() entity.Request {
	return i.Request.ToEntity()
}

func (i InspectVariables) LogValue() any {
	return map[string]any{"local_id": i.LocalId, "request": i.Request.LogValue()}
}
