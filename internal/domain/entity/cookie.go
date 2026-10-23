package entity

import (
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

func (c Cookie) MovedFrom(previous *Cookie) bool {
	return previous != nil && !c.Is(previous.Domain, previous.Path, previous.Name)
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

// MergeCookies puts the run's cookies over the workspace's: the same name, domain and path replace, others are added.
func MergeCookies(workspace, run []Cookie) []Cookie {
	merged := slices.Clone(workspace)
	for _, cookie := range run {
		index := slices.IndexFunc(merged, func(existing Cookie) bool { return existing.Is(cookie.Domain, cookie.Path, cookie.Name) })
		if index < 0 {
			merged = append(merged, cookie)
			continue
		}
		merged[index] = cookie
	}
	return merged
}
