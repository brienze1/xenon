package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type OAuthTokenRequest struct {
	Auth    *AuthOAuth2 `json:"auth,omitempty"`
	LocalId string      `json:"local_id" validate:"required,not_blank"`
	OwnerId string      `json:"owner_id" validate:"required,not_blank"`
}

func (o *OAuthTokenRequest) Draft() *entity.OAuth2Config {
	if o.Auth == nil {
		return nil
	}
	draft := o.Auth.ToEntity()
	return &draft
}

func (o OAuthTokenRequest) LogValue() any {
	value := map[string]any{"local_id": o.LocalId, "owner_id": o.OwnerId}
	if o.Auth != nil {
		value["auth"] = o.Auth.LogValue()
	}
	return value
}
