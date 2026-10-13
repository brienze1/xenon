package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type CookieDomainList []CookieDomain

func NewCookieDomainList(domains []entity.CookieDomain) CookieDomainList {
	list := make(CookieDomainList, 0, len(domains))
	for _, domain := range domains {
		list = append(list, NewCookieDomain(domain))
	}
	return list
}

func (c CookieDomainList) LogValue() any {
	return map[string]any{"domains": len(c)}
}
