package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type CookieJarReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
}

type CookieKey struct {
	Domain string `json:"domain"`
	Path   string `json:"path"`
	Name   string `json:"name"`
}

type UpsertCookie struct {
	Replaces *CookieKey `json:"replaces,omitempty"`
	LocalId  string     `json:"local_id" validate:"required,not_blank"`
	Domain   string     `json:"domain" validate:"required,not_blank"`
	Raw      string     `json:"raw" validate:"required,not_blank"`
}

type DeleteCookie struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Domain  string `json:"domain" validate:"required,not_blank"`
	Path    string `json:"path"`
	Name    string `json:"name" validate:"required"`
}

type DeleteCookieDomain struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Domain  string `json:"domain" validate:"required,not_blank"`
}

type SaveCookieAllowlist struct {
	LocalId string   `json:"local_id" validate:"required,not_blank"`
	Domains []string `json:"domains"`
}

type CookieAllowlist struct {
	Domains []string `json:"domains"`
}

func (u UpsertCookie) ReplacedCookie() *entity.Cookie {
	if u.Replaces == nil {
		return nil
	}
	domain := u.Replaces.Domain
	if domain == "" {
		domain = u.Domain
	}
	return &entity.Cookie{Domain: domain, Path: u.Replaces.Path, Name: u.Replaces.Name}
}

func (u UpsertCookie) LogValue() any {
	return map[string]any{"localId": u.LocalId, "domain": u.Domain, "replaces": u.Replaces}
}

func (d DeleteCookie) ToEntity() entity.Cookie {
	return entity.Cookie{Domain: d.Domain, Path: d.Path, Name: d.Name}
}

func NewCookieAllowlist(allowlist entity.CookieAllowlist) CookieAllowlist {
	return CookieAllowlist{Domains: append([]string{}, allowlist.Domains...)}
}
