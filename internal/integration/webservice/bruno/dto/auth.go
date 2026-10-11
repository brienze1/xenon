package dto

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
)

const (
	apiKeyPlacement  = "placement"
	apiKeyLocation   = "in"
	queryPlacement   = "queryparams"
	queryLocation    = "query"
	bearerTokenKey   = "token"
	basicUsernameKey = "username"
	basicPasswordKey = "password"
)

var attributeNames = map[string]map[string]string{
	authtype.OAuth2.String(): {
		"grant_type": "grant_type", "grantType": "grant_type",
		"access_token_url": "accessTokenUrl", "accessTokenUrl": "accessTokenUrl",
		"authorization_url": "authUrl", "authorizationUrl": "authUrl",
		"callback_url": "redirect_uri", "callbackUrl": "redirect_uri",
		"refresh_token_url": "refreshTokenUrl", "refreshTokenUrl": "refreshTokenUrl",
		"client_id": "clientId", "clientId": "clientId",
		"client_secret": "clientSecret", "clientSecret": "clientSecret",
		"credentials_placement": "client_authentication", "credentialsPlacement": "client_authentication",
	},
	authtype.AwsV4.String(): {
		"accessKeyId": "accessKey", "secretAccessKey": "secretKey",
	},
	authtype.ApiKey.String(): {
		apiKeyPlacement: apiKeyLocation,
	},
}

var mappedAuthTypes = []string{authtype.ApiKey.String(), authtype.OAuth2.String(), authtype.AwsV4.String(), authtype.Digest.String()}

type Auth struct {
	Mode   string
	Values []Pair
}

func NewAuth(mode string, values []Pair) Auth {
	return Auth{Mode: mode, Values: values}
}

func (a Auth) value(key string) string {
	for _, pair := range a.Values {
		if pair.Key == key {
			return pair.Value
		}
	}
	return ""
}

func (a Auth) ToEntity() entity.Auth {
	switch {
	case a.Mode == "" || a.Mode == authtype.Inherit.String():
		return entity.DefaultAuth()
	case a.Mode == authtype.None.String():
		return entity.Auth{Type: authtype.None}
	case a.Mode == authtype.Bearer.String():
		return entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: a.value(bearerTokenKey)}}
	case a.Mode == authtype.Basic.String():
		return entity.Auth{Type: authtype.Basic, Basic: entity.AuthBasic{Username: a.value(basicUsernameKey), Password: a.value(basicPasswordKey)}}
	case slices.Contains(mappedAuthTypes, a.Mode):
		return entity.NewAttributedAuth(authtype.GetAuthType(a.Mode), a.attributes())
	default:
		return entity.NewUnsupportedAuth(entity.AuthUnsupported{Origin: importorigin.Bruno, Type: a.Mode, Data: a.data()})
	}
}

func (a Auth) attributes() []entity.AuthAttribute {
	attributes := make([]entity.AuthAttribute, 0, len(a.Values))
	for _, pair := range a.Values {
		key, value := pair.Key, pair.Value
		if renamed, found := attributeNames[a.Mode][key]; found {
			key = renamed
		}
		if key == apiKeyLocation && value == queryPlacement {
			value = queryLocation
		}
		attributes = append(attributes, entity.AuthAttribute{Key: key, Value: value})
	}
	return attributes
}

func (a Auth) data() map[string]any {
	data := map[string]any{}
	for _, pair := range a.Values {
		data[pair.Key] = pair.Value
	}
	return data
}
