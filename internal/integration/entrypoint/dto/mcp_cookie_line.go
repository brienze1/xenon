package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPCookieLine struct {
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Expires  string `json:"expires,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HttpOnly bool   `json:"http_only,omitempty"`
}

func NewMCPCookieLine(cookie entity.Cookie) MCPCookieLine {
	converted := NewCookie(cookie)
	return MCPCookieLine{
		Domain:   converted.Domain,
		Name:     converted.Name,
		Value:    converted.Value,
		Path:     converted.Path,
		Expires:  converted.Expires,
		Secure:   converted.Secure,
		HttpOnly: converted.HttpOnly,
	}
}

func (m MCPCookieLine) LogValue() any {
	return map[string]any{"name": m.Name, "domain": m.Domain, "path": m.Path}
}
