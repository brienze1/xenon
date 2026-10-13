package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

func MergeImportedSecrets(existing, imported []entity.VariableRow, current, values map[string]string, resolution enums.ClashResolution) map[string]string {
	secrets := make(map[string]string, len(current))
	for key, value := range current {
		secrets[key] = value
	}
	for _, row := range imported {
		if _, clashes := entity.VariableRowIndex(existing, row.Key); clashes && resolution != clashresolution.Replace {
			continue
		}
		delete(secrets, row.Key)
		if value, found := values[row.Key]; found && row.Secret {
			secrets[row.Key] = value
		}
	}
	return secrets
}
