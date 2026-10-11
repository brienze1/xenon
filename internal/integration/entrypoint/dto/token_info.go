package dto

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type TokenInfo struct {
	ExpiresAt       *time.Time `json:"expires_at"`
	ObtainedAt      time.Time  `json:"obtained_at"`
	OwnerId         string     `json:"owner_id"`
	TokenType       string     `json:"token_type"`
	Scope           string     `json:"scope"`
	UseToken        string     `json:"use_token"`
	HasRefreshToken bool       `json:"has_refresh_token"`
	HasIdToken      bool       `json:"has_id_token"`
	Expired         bool       `json:"expired"`
}

func NewTokenInfo(token entity.OAuth2Token, now time.Time) TokenInfo {
	return TokenInfo{
		ExpiresAt:       token.ExpiresAt,
		ObtainedAt:      token.ObtainedAt,
		OwnerId:         token.Owner,
		TokenType:       token.TokenType,
		Scope:           token.Scope,
		UseToken:        token.UseToken,
		HasRefreshToken: token.CanRefresh(),
		HasIdToken:      token.HasIDToken(),
		Expired:         token.IsExpired(now),
	}
}
