package entity

import (
	"net/http"
	"strings"
	"time"
)

const setCookieHeader = "Set-Cookie"

type ScriptCookie struct {
	Name     string
	Value    string
	Domain   string
	Path     string
	Expires  string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	HostOnly bool
}

// NewScriptCookies reads the response's Set-Cookie headers; a cookie without a Domain belongs to the request host only.
func NewScriptCookies(headers []KeyValue, host string) []ScriptCookie {
	cookies := make([]ScriptCookie, 0)
	for _, header := range headers {
		if !header.HasKey(setCookieHeader) {
			continue
		}
		parsed, err := http.ParseSetCookie(header.Value)
		if err != nil {
			continue
		}
		cookie := ScriptCookie{Name: parsed.Name, Value: parsed.Value, Domain: strings.TrimPrefix(parsed.Domain, "."), Path: parsed.Path, MaxAge: parsed.MaxAge, Secure: parsed.Secure, HttpOnly: parsed.HttpOnly}
		if cookie.Domain == "" {
			cookie.Domain = host
			cookie.HostOnly = true
		}
		if !parsed.Expires.IsZero() {
			cookie.Expires = parsed.Expires.UTC().Format(time.RFC1123)
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}
