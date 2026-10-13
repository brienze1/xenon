package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type CookieDomain struct {
	Domain  string   `json:"domain"`
	Cookies []Cookie `json:"cookies"`
}

func NewCookieDomain(domain entity.CookieDomain) CookieDomain {
	return CookieDomain{Domain: domain.Domain, Cookies: NewCookies(domain.Cookies)}
}
