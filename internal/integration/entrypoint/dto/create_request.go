package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type CreateRequest struct {
	Request  *Request `json:"request,omitempty"`
	LocalId  string   `json:"local_id" validate:"required,not_blank"`
	ParentId string   `json:"parent_id" validate:"required,not_blank"`
	Name     string   `json:"name" validate:"required,not_blank,max=200"`
}

func (c *CreateRequest) RequestEntity() entity.Request {
	if c.Request == nil {
		return entity.NewRequest("", c.Name)
	}
	return c.Request.ToEntity()
}

func (c CreateRequest) LogValue() any {
	return map[string]any{"local_id": c.LocalId, "parent_id": c.ParentId, "name": c.Name}
}
