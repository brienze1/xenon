package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/vaultstate"
)

type Vault struct {
	Secrets      map[string]map[string]string
	OAuth2Tokens map[string]OAuth2Token
	LocalID      string
	Mode         enums.VaultMode
	State        enums.VaultState
	KeyAvailable bool
}

func NewVault(localId string, mode enums.VaultMode, state enums.VaultState, secrets map[string]map[string]string) Vault {
	return Vault{LocalID: localId, Mode: mode, State: state, Secrets: copyScopes(secrets)}
}

func NewVaultFromStorage(storage VaultStorage) Vault {
	vault := NewVault(storage.LocalID, storage.Mode, storage.State(), storage.Secrets)
	vault.KeyAvailable = storage.KeyAvailable
	vault.OAuth2Tokens = copyTokens(storage.OAuth2Tokens)
	return vault
}

func NewUnopenedVault(localId string) Vault {
	return Vault{LocalID: localId}
}

func (v Vault) IsEmpty() bool {
	return v.State == vaultstate.Empty
}

func (v Vault) IsLocked() bool {
	return v.State == vaultstate.Locked
}

func (v Vault) IsBroken() bool {
	return v.State == vaultstate.Broken
}

func (v Vault) IsReadable() bool {
	return v.State == vaultstate.Unlocked
}

func (v Vault) CanReveal() bool {
	return v.IsReadable() || v.IsEmpty()
}

func (v Vault) IsWritable() bool {
	return v.IsReadable() || v.IsEmpty() && v.KeyAvailable
}

func (v Vault) Unlocked() Vault {
	v.State = vaultstate.Unlocked
	return v
}

func (v Vault) SecretsFor(scope ScopeRef) map[string]string {
	if !v.IsReadable() {
		return map[string]string{}
	}
	return copyValues(v.Secrets[scope.String()])
}

func (v Vault) WithSecrets(scope ScopeRef, values map[string]string) Vault {
	v.Secrets = withScopeValues(v.Secrets, scope, values)
	return v
}

func (v Vault) WithoutScope(scope ScopeRef) Vault {
	return v.WithSecrets(scope, nil)
}

func (v Vault) HasScope(scope ScopeRef) bool {
	return len(v.SecretsFor(scope)) > 0
}

func (v Vault) TokenOf(owner string) (OAuth2Token, bool) {
	if !v.IsReadable() {
		return OAuth2Token{}, false
	}
	token, found := v.OAuth2Tokens[owner]
	return token, found
}

func (v Vault) WithToken(token OAuth2Token) Vault {
	v.OAuth2Tokens = copyTokens(v.OAuth2Tokens)
	v.OAuth2Tokens[token.Owner] = token
	return v
}

func (v Vault) WithTokens(tokens []OAuth2Token) Vault {
	for _, token := range tokens {
		v = v.WithToken(token)
	}
	return v
}

func (v Vault) WithoutToken(owner string) Vault {
	v.OAuth2Tokens = copyTokens(v.OAuth2Tokens)
	delete(v.OAuth2Tokens, owner)
	return v
}

func copyTokens(tokens map[string]OAuth2Token) map[string]OAuth2Token {
	copied := make(map[string]OAuth2Token, len(tokens))
	for owner, token := range tokens {
		copied[owner] = token
	}
	return copied
}

func (v Vault) LogValue() any {
	return map[string]any{"localId": v.LocalID, "mode": v.Mode, "state": v.State, "scopes": len(v.Secrets)}
}
