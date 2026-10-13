package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type Environment struct {
	Id                string        `json:"id"`
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	Color             string        `json:"color"`
	Variables         []VariableRow `json:"variables"`
	ConfirmBeforeSend bool          `json:"confirm_before_send"`
}

func NewEnvironment(details entity.EnvironmentDetails) Environment {
	return Environment{
		Id:                details.Environment.ID,
		Name:              details.Environment.Name,
		Description:       details.Environment.Description,
		Color:             details.Environment.Color.String(),
		ConfirmBeforeSend: details.Environment.ConfirmBeforeSend,
		Variables:         NewVariableRows(details.Table.Entries()),
	}
}
