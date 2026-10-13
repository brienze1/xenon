package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiSecurityScheme struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	Scheme           string   `json:"scheme"`
	In               string   `json:"in"`
	ParamName        string   `json:"param_name"`
	Flow             string   `json:"flow"`
	AuthorizationUrl string   `json:"authorization_url"`
	TokenUrl         string   `json:"token_url"`
	Description      string   `json:"description"`
	Scopes           []string `json:"scopes"`
}

func NewApiSecurityScheme(scheme entity.ApiSecurityScheme) ApiSecurityScheme {
	return ApiSecurityScheme{
		Name:             scheme.Name,
		Type:             scheme.Type,
		Scheme:           scheme.Scheme,
		In:               scheme.In,
		ParamName:        scheme.ParamName,
		Flow:             scheme.Flow,
		AuthorizationUrl: scheme.AuthorizationURL,
		TokenUrl:         scheme.TokenURL,
		Description:      scheme.Description,
		Scopes:           append([]string{}, scheme.Scopes...),
	}
}
