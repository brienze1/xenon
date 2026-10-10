package model

import "github.com/brienze1/xenon/internal/domain/entity"

const vaultContentVersion = 1

type VaultContent struct {
	Secrets map[string]map[string]string `json:"secrets"`
	Version int                          `json:"version"`
}

func NewVaultContent(vault entity.Vault) VaultContent {
	secrets := make(map[string]map[string]string, len(vault.Secrets))
	for scope, values := range vault.Secrets {
		if len(values) > 0 {
			secrets[scope] = values
		}
	}
	return VaultContent{Version: vaultContentVersion, Secrets: secrets}
}

func (v *VaultContent) ToEntity(storage entity.VaultStorage) entity.VaultStorage {
	storage.Secrets = v.Secrets
	return storage
}
