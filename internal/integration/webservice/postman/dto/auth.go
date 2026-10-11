package dto

import (
	"bytes"
	"encoding/json"
	"slices"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
)

const (
	noAuth      = "noauth"
	tokenKey    = "token"
	usernameKey = "username"
	passwordKey = "password"
)

var reservedAuthTypes = []string{authtype.ApiKey.String(), authtype.OAuth2.String(), authtype.AwsV4.String(), authtype.Digest.String()}

type Auth struct {
	Raw        map[string]any `json:"-"`
	Type       string         `json:"type"`
	Attributes []any          `json:"-"`
}

func (a *Auth) UnmarshalJSON(data []byte) error {
	type plain Auth
	if err := json.Unmarshal(data, (*plain)(a)); err != nil {
		return err
	}
	a.Raw = decodeRaw(data)
	var parts map[string]json.RawMessage
	_ = json.Unmarshal(data, &parts)
	a.Attributes = orderedAttributes(parts[a.Type])
	return nil
}

func (a *Auth) ToEntity() entity.Auth {
	if a == nil {
		return entity.DefaultAuth()
	}
	attributes := a.attributes()
	switch {
	case a.Type == noAuth:
		return entity.Auth{Type: authtype.None}
	case a.Type == authtype.Bearer.String():
		return entity.Auth{Type: authtype.Bearer, Bearer: entity.AuthBearer{Token: text(attributeValue(attributes, tokenKey))}}
	case a.Type == authtype.Basic.String():
		return entity.Auth{Type: authtype.Basic, Basic: entity.AuthBasic{Username: text(attributeValue(attributes, usernameKey)), Password: text(attributeValue(attributes, passwordKey))}}
	case slices.Contains(reservedAuthTypes, a.Type):
		return entity.NewAttributedAuth(authtype.GetAuthType(a.Type), authAttributes(attributes))
	default:
		return entity.NewUnsupportedAuth(entity.AuthUnsupported{Origin: importorigin.Postman, Type: a.Type, Data: attributes})
	}
}

func (a *Auth) IsUnsupported() bool {
	return a != nil && a.ToEntity().IsUnsupported()
}

func (a *Auth) Passthrough() map[string]any {
	if a == nil || a.IsUnsupported() {
		return map[string]any{}
	}
	types := map[string]any{}
	for _, attribute := range a.attributes() {
		if entry, isMap := attribute.(map[string]any); isMap && entry["type"] != nil {
			types[text(entry["key"])] = entry["type"]
		}
	}
	return withEntry(map[string]any{}, a.Type, types)
}

func (a *Auth) attributes() []any {
	return a.Attributes
}

func orderedAttributes(data json.RawMessage) []any {
	var list []any
	if err := json.Unmarshal(data, &list); err == nil {
		return list
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	attributes := make([]any, 0)
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return attributes
	}
	for decoder.More() {
		key, _ := decoder.Token()
		var value any
		_ = decoder.Decode(&value)
		attributes = append(attributes, map[string]any{"key": key, "value": value})
	}
	return attributes
}

func authAttributes(attributes []any) []entity.AuthAttribute {
	result := make([]entity.AuthAttribute, 0, len(attributes))
	for _, attribute := range attributes {
		if entry, isMap := attribute.(map[string]any); isMap {
			result = append(result, entity.AuthAttribute{Key: text(entry["key"]), Value: entry["value"]})
		}
	}
	return result
}

func attributeValue(attributes []any, key string) any {
	for _, attribute := range attributes {
		if entry, isMap := attribute.(map[string]any); isMap && text(entry["key"]) == key {
			return entry["value"]
		}
	}
	return nil
}
