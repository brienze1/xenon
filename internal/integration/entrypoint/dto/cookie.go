package dto

import (
	"net/http"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

var sameSiteModes = map[string]http.SameSite{"Lax": http.SameSiteLaxMode, "Strict": http.SameSiteStrictMode, "None": http.SameSiteNoneMode}

type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  string `json:"expires"`
	SameSite string `json:"same_site"`
	Raw      string `json:"raw"`
	Secure   bool   `json:"secure"`
	HttpOnly bool   `json:"http_only"`
	HostOnly bool   `json:"host_only"`
}

func NewCookie(cookie entity.Cookie) Cookie {
	line := http.Cookie{Name: cookie.Name, Value: cookie.Value, Path: cookie.Path, Secure: cookie.Secure, HttpOnly: cookie.HttpOnly, SameSite: sameSiteModes[cookie.SameSite]}
	converted := Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Domain:   cookie.Domain,
		Path:     cookie.Path,
		SameSite: cookie.SameSite,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
		HostOnly: cookie.HostOnly,
	}
	if !cookie.HostOnly {
		line.Domain = cookie.Domain
	}
	if !cookie.IsSession() {
		converted.Expires = cookie.Expires.UTC().Format(time.RFC3339)
		line.Expires = *cookie.Expires
	}
	converted.Raw = line.String()
	return converted
}

func NewCookies(cookies []entity.Cookie) []Cookie {
	converted := make([]Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		converted = append(converted, NewCookie(cookie))
	}
	return converted
}

func (c Cookie) LogValue() any {
	return map[string]any{"name": c.Name, "domain": c.Domain, "path": c.Path}
}
