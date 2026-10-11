package dto

import "github.com/brienze1/xenon/internal/domain/entity"

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
