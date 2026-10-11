package entity

import (
	"strings"
	"time"
)

const (
	tokenExpiryMargin = 30 * time.Second
	macTokenType      = "mac"
	tokenPartRefresh  = "refresh"
	tokenPartID       = "id"
)

type OAuth2Token struct {
	ExpiresAt    *time.Time
	ObtainedAt   time.Time
	AccessToken  string
	RefreshToken string
	IDToken      string
	TokenType    string
	Scope        string
	Owner        string
	UseToken     string
}

func (o OAuth2Token) IsExpired(now time.Time) bool {
	return o.ExpiresAt != nil && !o.ExpiresAt.After(now.Add(tokenExpiryMargin))
}

func (o OAuth2Token) CanRefresh() bool {
	return o.RefreshToken != ""
}

func (o OAuth2Token) HasIDToken() bool {
	return o.IDToken != ""
}

func (o OAuth2Token) IsMac() bool {
	return strings.EqualFold(o.TokenType, macTokenType)
}

func (o OAuth2Token) Value() string {
	if o.UseToken == UseIDToken {
		return o.IDToken
	}
	return o.AccessToken
}

func (o OAuth2Token) Part(which string) string {
	switch which {
	case tokenPartRefresh:
		return o.RefreshToken
	case tokenPartID:
		return o.IDToken
	default:
		return o.AccessToken
	}
}

func (o OAuth2Token) Secrets() []string {
	secrets := make([]string, 0, 3)
	for _, value := range []string{o.AccessToken, o.RefreshToken, o.IDToken} {
		if value != "" {
			secrets = append(secrets, value)
		}
	}
	return secrets
}

func (o *OAuth2Token) SecretsOrNone() []string {
	if o == nil {
		return []string{}
	}
	return o.Secrets()
}

func (o OAuth2Token) Using(useToken string) OAuth2Token {
	o.UseToken = useToken
	return o
}

func (o OAuth2Token) OwnedBy(owner, useToken string) OAuth2Token {
	o.Owner = owner
	o.UseToken = useToken
	return o
}

// Refreshed keeps the refresh and ID tokens the provider did not rotate.
func (o OAuth2Token) Refreshed(fresh OAuth2Token) OAuth2Token {
	if fresh.RefreshToken == "" {
		fresh.RefreshToken = o.RefreshToken
	}
	if fresh.IDToken == "" {
		fresh.IDToken = o.IDToken
	}
	if fresh.Scope == "" {
		fresh.Scope = o.Scope
	}
	return fresh.OwnedBy(o.Owner, o.UseToken)
}

func (o OAuth2Token) LogValue() any {
	return map[string]any{"owner": o.Owner, "tokenType": o.TokenType, "scope": o.Scope, "expiresAt": o.ExpiresAt, "hasRefreshToken": o.CanRefresh(), "hasIdToken": o.HasIDToken()}
}
