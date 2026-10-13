package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
)

type UpdateEnvironment struct {
	LocalId           string `json:"local_id" validate:"required,not_blank"`
	Id                string `json:"id" validate:"required,not_blank"`
	Name              string `json:"name" validate:"required,not_blank"`
	Description       string `json:"description"`
	Color             string `json:"color" validate:"oneof=none green cyan magenta yellow red"`
	ConfirmBeforeSend bool   `json:"confirm_before_send"`
}

func (u *UpdateEnvironment) ToEntity() entity.Environment {
	return entity.Environment{
		ID:                u.Id,
		Name:              strings.TrimSpace(u.Name),
		Description:       u.Description,
		Color:             envcolor.GetEnvColor(u.Color),
		ConfirmBeforeSend: u.ConfirmBeforeSend,
	}
}
