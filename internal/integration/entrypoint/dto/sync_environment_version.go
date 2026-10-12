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

type SyncVariableVersion struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Secret      bool   `json:"secret"`
	Enabled     bool   `json:"enabled"`
}

func NewSyncEnvironmentVersion(environment entity.Environment) SyncEnvironmentVersion {
	variables := make([]SyncVariableVersion, 0, len(environment.Variables))
	for _, row := range environment.Variables {
		stored := row.WithoutSecretValue()
		variables = append(variables, SyncVariableVersion{Key: stored.Key, Value: stored.InitialValue, Description: stored.Description, Secret: stored.Secret, Enabled: stored.Enabled})
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
