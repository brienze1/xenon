package entity

import (
	"cmp"
	"slices"
	"strings"
	"time"
)

const HeaderCookie = "Cookie"

type Cookie struct {
	Expires    *time.Time
	Created    time.Time
	LastAccess time.Time
	Name       string
	Value      string
	Domain     string
	Path       string
	SameSite   string
	HostOnly   bool
	Secure     bool
	HttpOnly   bool
}

func (c Cookie) IsSession() bool {
	return c.Expires == nil
}

func (c Cookie) IsOn(domain string) bool {
	return strings.EqualFold(c.Domain, domain)
}

func (c Cookie) Is(domain, path, name string) bool {
	return c.IsOn(domain) && c.Path == path && c.Name == name
}

func (c Cookie) LogValue() any {
	return map[string]any{"name": c.Name, "domain": c.Domain, "path": c.Path}
}

func PersistentCookies(cookies []Cookie) []Cookie {
	return slices.DeleteFunc(slices.Clone(cookies), Cookie.IsSession)
}

func CookiesOn(cookies []Cookie, domain string) []Cookie {
	return slices.DeleteFunc(slices.Clone(cookies), func(cookie Cookie) bool { return !cookie.IsOn(domain) })
}

func FirstCookieNamed(cookies []Cookie, name string) []Cookie {
	index := slices.IndexFunc(cookies, func(cookie Cookie) bool { return cookie.Name == name })
	if index < 0 {
		return []Cookie{}
	}
	return cookies[index : index+1]
}

func CookieHeader(cookies []Cookie) string {
	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		pairs = append(pairs, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(pairs, "; ")
}

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

type CookieJarState struct {
	Cookies       []Cookie
	Revision      uint64
	SavedRevision uint64
	Loaded        bool
}

func (c CookieJarState) HasUnsavedChanges() bool {
	return c.Revision != c.SavedRevision
}

type StoredCookies struct {
	Cookies   []Cookie
	Available bool
}

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
