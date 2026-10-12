package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncVariableVersion struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Secret      bool   `json:"secret"`
	Enabled     bool   `json:"enabled"`
}

func NewSyncVariableVersion(row entity.VariableRow) SyncVariableVersion {
	stored := row.WithoutSecretValue()
	return SyncVariableVersion{
		Key:         stored.Key,
		Value:       stored.InitialValue,
		Description: stored.Description,
		Secret:      stored.Secret,
		Enabled:     stored.Enabled,
	}
}
