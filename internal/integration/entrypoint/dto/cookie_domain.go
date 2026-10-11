package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type CookieDomain struct {
	Domain  string   `json:"domain"`
	Cookies []Cookie `json:"cookies"`
}

type CookieDomainList []CookieDomain

func NewCookieDomainList(domains []entity.CookieDomain) CookieDomainList {
	list := make(CookieDomainList, 0, len(domains))
	for _, domain := range domains {
		list = append(list, CookieDomain{Domain: domain.Domain, Cookies: NewCookies(domain.Cookies)})
	}
	return list
}

func (c CookieDomainList) LogValue() any {
	return map[string]any{"domains": len(c)}
}
