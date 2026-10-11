package dto

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type OAuth2TokenResponse struct {
	ExpiresIn        json.Number `json:"expires_in"`
	AccessToken      string      `json:"access_token"`
	RefreshToken     string      `json:"refresh_token"`
	IdToken          string      `json:"id_token"`
	TokenType        string      `json:"token_type"`
	Scope            string      `json:"scope"`
	Error            string      `json:"error"`
	ErrorDescription string      `json:"error_description"`
}

// NewOAuth2TokenResponse reads a JSON or form-encoded token endpoint answer.
func NewOAuth2TokenResponse(body []byte) OAuth2TokenResponse {
	var response OAuth2TokenResponse
	if json.Unmarshal(body, &response) == nil {
		return response
	}
	form, err := url.ParseQuery(string(body))
	if err != nil {
		return response
	}
	return OAuth2TokenResponse{
		ExpiresIn:        json.Number(form.Get("expires_in")),
		AccessToken:      form.Get("access_token"),
		RefreshToken:     form.Get("refresh_token"),
		IdToken:          form.Get("id_token"),
		TokenType:        form.Get("token_type"),
		Scope:            form.Get("scope"),
		Error:            form.Get("error"),
		ErrorDescription: form.Get("error_description"),
	}
}

func (o OAuth2TokenResponse) HasError() bool {
	return o.Error != "" || o.AccessToken == ""
}

func (o OAuth2TokenResponse) ErrorCode(status int) string {
	if o.Error != "" {
		return o.Error
	}
	return "http_" + strconv.Itoa(status)
}

func (o OAuth2TokenResponse) ToEntity(now time.Time, requestedScope string) entity.OAuth2Token {
	token := entity.OAuth2Token{AccessToken: o.AccessToken, RefreshToken: o.RefreshToken, IDToken: o.IdToken, TokenType: o.TokenType, Scope: o.Scope, ObtainedAt: now}
	if token.Scope == "" {
		token.Scope = requestedScope
	}
	if seconds, err := o.ExpiresIn.Int64(); err == nil {
		expiresAt := now.Add(time.Duration(seconds) * time.Second)
		token.ExpiresAt = &expiresAt
	}
	return token
}
