package entity

import (
	"net/url"
	"slices"
	"strings"
)

type EffectiveRequest struct {
	AuthOwner           string
	CookieJar           string
	Method              string
	URL                 string
	Body                Body
	Headers             []KeyValue
	AutoHeaders         []AutoHeader
	DisabledAutoHeaders []string
	Auth                Auth
	Settings            RequestSettings
}

func (e EffectiveRequest) HasHeader(name string) bool {
	for _, header := range e.Headers {
		if header.HasKey(name) {
			return true
		}
	}
	return false
}

func (e EffectiveRequest) IsAutoHeaderDisabled(name string) bool {
	for _, disabled := range e.DisabledAutoHeaders {
		if strings.EqualFold(strings.TrimSpace(disabled), name) {
			return true
		}
	}
	return false
}

func (e EffectiveRequest) WithAutoHeaders(headers []AutoHeader) EffectiveRequest {
	e.AutoHeaders = headers
	return e
}

func (e EffectiveRequest) UsesAutoHeader(name string) bool {
	for _, header := range e.AutoHeaders {
		if strings.EqualFold(header.Key, name) {
			return header.IsActive()
		}
	}
	return false
}

func (e EffectiveRequest) SendHeaders() []KeyValue {
	headers := append([]KeyValue{}, e.Headers...)
	for _, header := range e.AutoHeaders {
		if header.IsSent() {
			headers = append(headers, NewKeyValue(header.Key, header.Value))
		}
	}
	return headers
}

func (e EffectiveRequest) TimeoutText() string {
	return e.Settings.TimeoutText()
}

func (e EffectiveRequest) LogValue() any {
	return map[string]any{"method": e.Method, "url": e.URL}
}

func (e EffectiveRequest) Host() string {
	parsed, err := url.Parse(e.URL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func (e EffectiveRequest) WithoutHeader(name string) EffectiveRequest {
	e.Headers = slices.DeleteFunc(slices.Clone(e.Headers), func(header KeyValue) bool {
		return header.HasKey(name)
	})
	e.AutoHeaders = slices.DeleteFunc(slices.Clone(e.AutoHeaders), func(header AutoHeader) bool {
		return strings.EqualFold(header.Key, name)
	})
	return e
}

func (e EffectiveRequest) WithHeader(name, value string) EffectiveRequest {
	e = e.WithoutHeader(name)
	e.Headers = append(e.Headers, NewKeyValue(name, value))
	return e
}

func (e EffectiveRequest) WithoutQueryParam(name string) EffectiveRequest {
	e.URL = ParseRequestUrl(e.URL).withoutQueryKey(name).String()
	return e
}

func (e EffectiveRequest) WithQueryParam(name, value string) EffectiveRequest {
	e.URL = ParseRequestUrl(e.URL).withoutQueryKey(name).withQueryPair(name, value).String()
	return e
}

func (e EffectiveRequest) WithURL(raw string) EffectiveRequest {
	e.URL = raw
	return e
}

func (e EffectiveRequest) WithCookieJar(localId string) EffectiveRequest {
	e.CookieJar = localId
	return e
}

func (e EffectiveRequest) UsesCookieJar() bool {
	return e.CookieJar != "" && !e.Settings.DisableCookieJar
}

func (e EffectiveRequest) SendsJarCookies() bool {
	return !e.IsAutoHeaderDisabled(HeaderCookie)
}

// ForAlias moves an alias's own auth to the alias's owner id, so its OAuth 2.0 token is not the Default's.
func (e EffectiveRequest) ForAlias(requestId, aliasId string) EffectiveRequest {
	if aliasId != "" && e.AuthOwner == requestId {
		e.AuthOwner = AliasAuthOwner(requestId, aliasId)
	}
	return e
}

func (e EffectiveRequest) WithAuthOwner(owner string) EffectiveRequest {
	if e.AuthOwner == "" {
		e.AuthOwner = owner
	}
	return e
}
