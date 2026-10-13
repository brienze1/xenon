package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type EnvironmentSummary struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Color             string `json:"color"`
	Error             string `json:"error,omitempty"`
	ConfirmBeforeSend bool   `json:"confirm_before_send"`
}

func NewEnvironmentSummary(environment entity.Environment) EnvironmentSummary {
	return EnvironmentSummary{
		Id:                environment.ID,
		Name:              environment.Name,
		Color:             environment.Color.String(),
		Error:             environment.Error,
		ConfirmBeforeSend: environment.ConfirmBeforeSend,
	}
}
