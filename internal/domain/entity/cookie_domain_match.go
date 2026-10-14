package entity

import "strings"

// Matches tells whether the domain is the one a filter names; an empty filter matches every domain.
func (c CookieDomain) Matches(filter string) bool {
	return filter == "" || strings.EqualFold(c.Domain, filter)
}

func CookiesMatching(domains []CookieDomain, filter string) []Cookie {
	cookies := make([]Cookie, 0)
	for _, domain := range domains {
		if domain.Matches(filter) {
			cookies = append(cookies, domain.Cookies...)
		}
	}
	return cookies
}
