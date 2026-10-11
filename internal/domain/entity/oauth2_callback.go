package entity

import (
	"strconv"
	"time"
)

type OAuth2Callback struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
	AccessToken      string
	TokenType        string
	ExpiresIn        string
	Scope            string
	IDToken          string
}

func (o OAuth2Callback) IsError() bool {
	return o.Error != ""
}

func (o OAuth2Callback) MatchesState(state string) bool {
	return o.State == state
}

func (o OAuth2Callback) Token(now time.Time, scope string) OAuth2Token {
	token := OAuth2Token{AccessToken: o.AccessToken, TokenType: o.TokenType, Scope: o.Scope, IDToken: o.IDToken, ObtainedAt: now}
	if token.Scope == "" {
		token.Scope = scope
	}
	if seconds, err := strconv.Atoi(o.ExpiresIn); err == nil {
		expiresAt := now.Add(time.Duration(seconds) * time.Second)
		token.ExpiresAt = &expiresAt
	}
	return token
}

func (o OAuth2Callback) LogValue() any {
	return map[string]any{"hasCode": o.Code != "", "hasAccessToken": o.AccessToken != "", "error": o.Error}
}
