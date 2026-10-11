package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AuthApiKey struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	In    string `json:"in" validate:"omitempty,oneof=header query"`
}

func NewAuthApiKey(apiKey entity.ApiKeyAuth) *AuthApiKey {
	return &AuthApiKey{Key: apiKey.Key, Value: apiKey.Value, In: apiKey.In}
}

func (a *AuthApiKey) ToEntity() entity.ApiKeyAuth {
	if a == nil {
		return entity.ApiKeyAuth{}
	}
	apiKey := entity.DefaultApiKeyAuth()
	apiKey.Key = a.Key
	apiKey.Value = a.Value
	if a.In == entity.ApiKeyInQuery {
		apiKey.In = entity.ApiKeyInQuery
	}
	return apiKey
}
