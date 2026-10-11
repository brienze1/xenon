package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type AuthApiKey struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
	In    string `yaml:"in" json:"in"`
}

func NewAuthApiKey(apiKey entity.ApiKeyAuth) AuthApiKey {
	return AuthApiKey{Key: apiKey.Key, Value: apiKey.Value, In: apiKey.In}
}

func (a *AuthApiKey) ToEntity() entity.ApiKeyAuth {
	apiKey := entity.DefaultApiKeyAuth()
	apiKey.Key = a.Key
	apiKey.Value = a.Value
	if a.In == entity.ApiKeyInQuery {
		apiKey.In = entity.ApiKeyInQuery
	}
	return apiKey
}

func (a *AuthApiKey) fields() []yamlnode.Field {
	apiKey := a.ToEntity()
	return []yamlnode.Field{yamlnode.String("key", apiKey.Key), yamlnode.String("value", apiKey.Value), yamlnode.String("in", apiKey.In)}
}
