package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type CookieKey struct {
	Domain string `json:"domain"`
	Path   string `json:"path"`
	Name   string `json:"name"`
}

func (k CookieKey) ToEntity(fallbackDomain string) entity.Cookie {
	domain := k.Domain
	if domain == "" {
		domain = fallbackDomain
	}
	return entity.Cookie{Domain: domain, Path: k.Path, Name: k.Name}
}
