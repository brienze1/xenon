package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncEnvironmentVersion struct {
	Id                string                `json:"id"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	Color             string                `json:"color"`
	Variables         []SyncVariableVersion `json:"variables"`
	ConfirmBeforeSend bool                  `json:"confirm_before_send"`
}

func NewSyncEnvironmentVersion(environment entity.Environment) SyncEnvironmentVersion {
	variables := make([]SyncVariableVersion, 0, len(environment.Variables))
	for _, row := range environment.Variables {
		variables = append(variables, NewSyncVariableVersion(row))
	}
	return SyncEnvironmentVersion{
		Id:                environment.ID,
		Name:              environment.Name,
		Description:       environment.Description,
		Color:             environment.Color.String(),
		Variables:         variables,
		ConfirmBeforeSend: environment.ConfirmBeforeSend,
	}
}
