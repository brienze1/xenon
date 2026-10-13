package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type OAuth2Token struct {
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	ObtainedAt   time.Time  `json:"obtained_at"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	IdToken      string     `json:"id_token,omitempty"`
	TokenType    string     `json:"token_type"`
	Scope        string     `json:"scope,omitempty"`
	UseToken     string     `json:"use_token"`
}

func NewOAuth2Token(token entity.OAuth2Token) OAuth2Token {
	return OAuth2Token{ExpiresAt: token.ExpiresAt, ObtainedAt: token.ObtainedAt, AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, IdToken: token.IDToken, TokenType: token.TokenType, Scope: token.Scope, UseToken: token.UseToken}
}

func (o *OAuth2Token) ToEntity(owner string) entity.OAuth2Token {
	return entity.OAuth2Token{ExpiresAt: o.ExpiresAt, ObtainedAt: o.ObtainedAt, AccessToken: o.AccessToken, RefreshToken: o.RefreshToken, IDToken: o.IdToken, TokenType: o.TokenType, Scope: o.Scope, Owner: owner, UseToken: o.UseToken}
}
