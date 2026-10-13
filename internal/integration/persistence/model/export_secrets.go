package model

import "github.com/brienze1/xenon/internal/domain/entity"

const exportSecretsVersion = 1

type ExportSecrets struct {
	Secrets map[string]map[string]string `json:"secrets"`
	Version int                          `json:"version"`
}

func NewExportSecrets(bundle entity.SecretBundle) ExportSecrets {
	return ExportSecrets{Version: exportSecretsVersion, Secrets: bundle.Scopes}
}

func (e *ExportSecrets) IsNewer() bool {
	return e.Version > exportSecretsVersion
}

func (e *ExportSecrets) ToEntity() entity.SecretBundle {
	return entity.NewSecretBundle(e.Secrets)
}
