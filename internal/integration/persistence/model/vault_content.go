package model

import "github.com/brienze1/xenon/internal/domain/entity"

const vaultContentVersion = 1

type VaultContent struct {
	Secrets map[string]map[string]string `json:"secrets"`
	OAuth2  map[string]OAuth2Token       `json:"oauth2,omitempty"`
	Version int                          `json:"version"`
}

func NewVaultContent(vault entity.Vault) VaultContent {
	secrets := make(map[string]map[string]string, len(vault.Secrets))
	for scope, values := range vault.Secrets {
		if len(values) > 0 {
			secrets[scope] = values
		}
	}
	tokens := make(map[string]OAuth2Token, len(vault.OAuth2Tokens))
	for owner, token := range vault.OAuth2Tokens {
		tokens[owner] = NewOAuth2Token(token)
	}
	return VaultContent{Version: vaultContentVersion, Secrets: secrets, OAuth2: tokens}
}

func (v *VaultContent) ToEntity(storage entity.VaultStorage) entity.VaultStorage {
	storage.Secrets = v.Secrets
	storage.OAuth2Tokens = make(map[string]entity.OAuth2Token, len(v.OAuth2))
	for owner, token := range v.OAuth2 {
		storage.OAuth2Tokens[owner] = token.ToEntity(owner)
	}
	return storage
}
