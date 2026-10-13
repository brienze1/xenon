package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type CookieAllowlist struct {
	Domains []string `json:"domains"`
}

func NewCookieAllowlist(allowlist entity.CookieAllowlist) CookieAllowlist {
	return CookieAllowlist{Domains: append([]string{}, allowlist.Domains...)}
}
