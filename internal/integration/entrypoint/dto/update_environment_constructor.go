package dto

import "github.com/brienze1/xenon/internal/domain/entity"

// NewUpdateEnvironment keeps every stored setting the edit leaves out.
func NewUpdateEnvironment(localId string, stored entity.Environment, edit *MCPEnvEdit) UpdateEnvironment {
	return UpdateEnvironment{
		LocalId:           localId,
		Id:                stored.ID,
		Name:              overridden(edit.Name, stored.Name),
		Description:       stored.Description,
		Color:             overridden(edit.Color, stored.Color.String()),
		ConfirmBeforeSend: overridden(edit.ConfirmBeforeSend, stored.ConfirmBeforeSend),
	}
}
