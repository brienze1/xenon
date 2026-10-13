package entity

import (
	"slices"
	"strings"
)

type CookieAllowlist struct {
	Domains []string
}

func NewCookieAllowlist(domains []string) CookieAllowlist {
	normalized := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain != "" && !slices.Contains(normalized, domain) {
			normalized = append(normalized, domain)
		}
	}
	return CookieAllowlist{Domains: normalized}
}

func (c CookieAllowlist) Allows(host string) bool {
	host = strings.ToLower(host)
	return slices.ContainsFunc(c.Domains, func(domain string) bool {
		return host == domain || strings.HasSuffix(host, "."+domain)
	})
}
