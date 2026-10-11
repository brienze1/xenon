package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/clientauthentication"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/tokenplacement"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type AuthOAuth2 struct {
	HeaderPrefix         *string       `yaml:"headerPrefix" json:"headerPrefix,omitempty"`
	AutoRefresh          *bool         `yaml:"autoRefresh" json:"autoRefresh,omitempty"`
	GrantType            string        `yaml:"grant_type" json:"grant_type"`
	TokenName            string        `yaml:"tokenName" json:"tokenName"`
	RedirectUri          string        `yaml:"redirect_uri" json:"redirect_uri"`
	AuthUrl              string        `yaml:"authUrl" json:"authUrl"`
	AccessTokenUrl       string        `yaml:"accessTokenUrl" json:"accessTokenUrl"`
	RefreshTokenUrl      string        `yaml:"refreshTokenUrl" json:"refreshTokenUrl"`
	ClientId             string        `yaml:"clientId" json:"clientId"`
	ClientSecret         string        `yaml:"clientSecret" json:"clientSecret"`
	Username             string        `yaml:"username" json:"username"`
	Password             string        `yaml:"password" json:"password"`
	Scope                string        `yaml:"scope" json:"scope"`
	State                string        `yaml:"state" json:"state"`
	ChallengeAlgorithm   string        `yaml:"challengeAlgorithm" json:"challengeAlgorithm"`
	CodeVerifier         string        `yaml:"code_verifier" json:"code_verifier"`
	ClientAuthentication string        `yaml:"client_authentication" json:"client_authentication"`
	AddTokenTo           string        `yaml:"addTokenTo" json:"addTokenTo"`
	UseToken             string        `yaml:"useToken" json:"useToken"`
	Audience             string        `yaml:"audience" json:"audience"`
	Resource             string        `yaml:"resource" json:"resource"`
	AuthRequestParams    []OAuth2Param `yaml:"authRequestParams" json:"authRequestParams"`
	TokenRequestParams   []OAuth2Param `yaml:"tokenRequestParams" json:"tokenRequestParams"`
	RefreshRequestParams []OAuth2Param `yaml:"refreshRequestParams" json:"refreshRequestParams"`
}

func NewAuthOAuth2(config entity.OAuth2Config) AuthOAuth2 {
	return AuthOAuth2{
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
		HeaderPrefix:         &config.HeaderPrefix,
		UseToken:             config.UseToken,
		AutoRefresh:          &config.AutoRefresh,
		AuthRequestParams:    newOAuth2Params(config.AuthRequestParams),
		TokenRequestParams:   newOAuth2Params(config.TokenRequestParams),
		RefreshRequestParams: newOAuth2Params(config.RefreshRequestParams),
		Audience:             config.Audience,
		Resource:             config.Resource,
	}
}

func (a *AuthOAuth2) ToEntity() entity.OAuth2Config {
	config := entity.DefaultOAuth2Config()
	if grant := oauth2grant.GetOAuth2Grant(a.GrantType); grant != oauth2grant.Unknown {
		config.GrantType = grant
	}
	config.TokenName = a.TokenName
	config.RedirectURI = withDefaultText(a.RedirectUri, config.RedirectURI)
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
	if a.HeaderPrefix != nil {
		config.HeaderPrefix = *a.HeaderPrefix
	}
	if a.UseToken == entity.UseIDToken {
		config.UseToken = entity.UseIDToken
	}
	if a.AutoRefresh != nil {
		config.AutoRefresh = *a.AutoRefresh
	}
	config.AuthRequestParams = oauth2ParamsToEntity(a.AuthRequestParams, entity.OAuth2SendAsUrl)
	config.TokenRequestParams = oauth2ParamsToEntity(a.TokenRequestParams, entity.OAuth2SendAsBody)
	config.RefreshRequestParams = oauth2ParamsToEntity(a.RefreshRequestParams, entity.OAuth2SendAsBody)
	config.Audience = a.Audience
	config.Resource = a.Resource
	return config
}

func (a *AuthOAuth2) fields() []yamlnode.Field {
	config := a.ToEntity()
	return []yamlnode.Field{
		yamlnode.String("grant_type", config.GrantType.String()),
		yamlnode.String("tokenName", config.TokenName),
		yamlnode.String("redirect_uri", config.RedirectURI),
		yamlnode.String("authUrl", config.AuthURL),
		yamlnode.String("accessTokenUrl", config.AccessTokenURL),
		yamlnode.String("refreshTokenUrl", config.RefreshTokenURL),
		yamlnode.String("clientId", config.ClientID),
		yamlnode.String("clientSecret", config.ClientSecret),
		yamlnode.String("username", config.Username),
		yamlnode.String("password", config.Password),
		yamlnode.String("scope", config.Scope),
		yamlnode.String("state", config.State),
		yamlnode.String("challengeAlgorithm", config.ChallengeAlgorithm),
		yamlnode.String("code_verifier", config.CodeVerifier),
		yamlnode.String("client_authentication", config.ClientAuthentication.String()),
		yamlnode.String("addTokenTo", config.AddTokenTo.String()),
		yamlnode.String("headerPrefix", config.HeaderPrefix),
		yamlnode.String("useToken", config.UseToken),
		yamlnode.Bool("autoRefresh", config.AutoRefresh),
		yamlnode.List("authRequestParams", oauth2ParamNodes(config.AuthRequestParams)),
		yamlnode.List("tokenRequestParams", oauth2ParamNodes(config.TokenRequestParams)),
		yamlnode.List("refreshRequestParams", oauth2ParamNodes(config.RefreshRequestParams)),
		yamlnode.String("audience", config.Audience),
		yamlnode.String("resource", config.Resource),
	}
}

func withDefaultText(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
