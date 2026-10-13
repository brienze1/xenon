package dto

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ScriptSandboxCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Expires  string `json:"expires,omitempty"`
	MaxAge   int    `json:"maxAge,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
	HostOnly bool   `json:"hostOnly,omitempty"`
}

func NewScriptSandboxCookies(cookies []entity.ScriptCookie) []ScriptSandboxCookie {
	converted := make([]ScriptSandboxCookie, 0, len(cookies))
	for _, cookie := range cookies {
		converted = append(converted, ScriptSandboxCookie(cookie))
	}
	return converted
}

func NewScriptSandboxJarCookie(cookie entity.Cookie) ScriptSandboxCookie {
	converted := ScriptSandboxCookie{Name: cookie.Name, Value: cookie.Value, Domain: cookie.Domain, Path: cookie.Path, Secure: cookie.Secure, HttpOnly: cookie.HttpOnly, HostOnly: cookie.HostOnly}
	if !cookie.IsSession() {
		converted.Expires = cookie.Expires.UTC().Format(time.RFC3339)
	}
	return converted
}

func NewScriptSandboxJarCookies(cookies []entity.Cookie) []ScriptSandboxCookie {
	converted := make([]ScriptSandboxCookie, 0, len(cookies))
	for _, cookie := range cookies {
		converted = append(converted, NewScriptSandboxJarCookie(cookie))
	}
	return converted
}
