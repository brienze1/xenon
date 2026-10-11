package entity

import (
	"cmp"
	"slices"
)

type CookieDomain struct {
	Domain  string
	Cookies []Cookie
}

func GroupCookiesByDomain(cookies []Cookie) []CookieDomain {
	groups := make([]CookieDomain, 0)
	sorted := slices.Clone(cookies)
	slices.SortFunc(sorted, func(a, b Cookie) int {
		return cmp.Or(cmp.Compare(a.Domain, b.Domain), cmp.Compare(a.Name, b.Name), cmp.Compare(a.Path, b.Path))
	})
	for _, cookie := range sorted {
		last := len(groups) - 1
		if last < 0 || groups[last].Domain != cookie.Domain {
			groups = append(groups, CookieDomain{Domain: cookie.Domain, Cookies: []Cookie{}})
			last++
		}
		groups[last].Cookies = append(groups[last].Cookies, cookie)
	}
	return groups
}
