package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/clientauthentication"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/tokenplacement"
)

type AuthOAuth2 struct {
	GrantType            string        `json:"grant_type" validate:"omitempty,oneof=authorization_code authorization_code_with_pkce implicit password_credentials client_credentials"`
	TokenName            string        `json:"token_name"`
	RedirectUri          string        `json:"redirect_uri"`
	AuthUrl              string        `json:"auth_url"`
	AccessTokenUrl       string        `json:"access_token_url"`
	RefreshTokenUrl      string        `json:"refresh_token_url"`
	ClientId             string        `json:"client_id"`
	ClientSecret         string        `json:"client_secret"`
	Username             string        `json:"username"`
	Password             string        `json:"password"`
	Scope                string        `json:"scope"`
	State                string        `json:"state"`
	ChallengeAlgorithm   string        `json:"challenge_algorithm" validate:"omitempty,oneof=S256 plain"`
	CodeVerifier         string        `json:"code_verifier"`
	ClientAuthentication string        `json:"client_authentication" validate:"omitempty,oneof=header body"`
	AddTokenTo           string        `json:"add_token_to" validate:"omitempty,oneof=header queryParams"`
	HeaderPrefix         string        `json:"header_prefix"`
	UseToken             string        `json:"use_token" validate:"omitempty,oneof=access_token id_token"`
	Audience             string        `json:"audience"`
	Resource             string        `json:"resource"`
	SuggestedCallback    string        `json:"suggested_callback,omitempty"`
	AuthRequestParams    []OAuth2Param `json:"auth_request_params" validate:"dive"`
	TokenRequestParams   []OAuth2Param `json:"token_request_params" validate:"dive"`
	RefreshRequestParams []OAuth2Param `json:"refresh_request_params" validate:"dive"`
	AutoRefresh          bool          `json:"auto_refresh"`
	Legacy               bool          `json:"legacy,omitempty"`
	UsesPostmanCallback  bool          `json:"uses_postman_callback,omitempty"`
}

func NewAuthOAuth2(config entity.OAuth2Config) *AuthOAuth2 {
	result := &AuthOAuth2{
		GrantType:            config.GrantType.String(),
		TokenName:            config.TokenName,
		RedirectUri:          config.RedirectURI,
		AuthUrl:              config.AuthURL,
		AccessTokenUrl:       config.AccessTokenURL,
		RefreshTokenUrl:      config.RefreshTokenURL,
		ClientId:             config.ClientID,
		ClientSecret:         config.ClientSecret,
		Username:             config.Username,
		Password:             config.Password,
		Scope:                config.Scope,
		State:                config.State,
		ChallengeAlgorithm:   config.ChallengeAlgorithm,
		CodeVerifier:         config.CodeVerifier,
		ClientAuthentication: config.ClientAuthentication.String(),
		AddTokenTo:           config.AddTokenTo.String(),
		HeaderPrefix:         config.HeaderPrefix,
		UseToken:             config.UseToken,
		Audience:             config.Audience,
		Resource:             config.Resource,
		AuthRequestParams:    NewOAuth2Params(config.AuthRequestParams),
		TokenRequestParams:   NewOAuth2Params(config.TokenRequestParams),
		RefreshRequestParams: NewOAuth2Params(config.RefreshRequestParams),
		AutoRefresh:          config.AutoRefresh,
		Legacy:               config.IsLegacyGrant(),
		UsesPostmanCallback:  config.UsesPostmanCallback(),
	}
	if result.UsesPostmanCallback {
		result.SuggestedCallback = entity.DefaultOAuth2Callback
	}
	return result
}

func (a *AuthOAuth2) ToEntity() entity.OAuth2Config {
	if a == nil {
		return entity.OAuth2Config{}
	}
	config := entity.DefaultOAuth2Config()
	if grant := oauth2grant.GetOAuth2Grant(a.GrantType); grant != oauth2grant.Unknown {
		config.GrantType = grant
	}
	config.TokenName = a.TokenName
	if a.RedirectUri != "" {
		config.RedirectURI = a.RedirectUri
	}
	config.AuthURL = a.AuthUrl
	config.AccessTokenURL = a.AccessTokenUrl
	config.RefreshTokenURL = a.RefreshTokenUrl
	config.ClientID = a.ClientId
	config.ClientSecret = a.ClientSecret
	config.Username = a.Username
	config.Password = a.Password
	config.Scope = a.Scope
	config.State = a.State
	if a.ChallengeAlgorithm == entity.PkcePlain {
		config.ChallengeAlgorithm = entity.PkcePlain
	}
	config.CodeVerifier = a.CodeVerifier
	if clientauthentication.Body.Equals(a.ClientAuthentication) {
		config.ClientAuthentication = clientauthentication.Body
	}
	if tokenplacement.QueryParams.Equals(a.AddTokenTo) {
		config.AddTokenTo = tokenplacement.QueryParams
	}
	config.HeaderPrefix = a.HeaderPrefix
	if a.UseToken == entity.UseIDToken {
		config.UseToken = entity.UseIDToken
	}
	config.Audience = a.Audience
	config.Resource = a.Resource
	config.AuthRequestParams = oauth2ParamsToEntity(a.AuthRequestParams, entity.OAuth2SendAsUrl)
	config.TokenRequestParams = oauth2ParamsToEntity(a.TokenRequestParams, entity.OAuth2SendAsBody)
	config.RefreshRequestParams = oauth2ParamsToEntity(a.RefreshRequestParams, entity.OAuth2SendAsBody)
	config.AutoRefresh = a.AutoRefresh
	return config
}

func (a AuthOAuth2) LogValue() any {
	return map[string]any{"grant_type": a.GrantType, "auth_url": a.AuthUrl, "access_token_url": a.AccessTokenUrl, "client_id": a.ClientId, "has_client_secret": a.ClientSecret != ""}
}
