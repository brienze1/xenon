package dto

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/cookiejar"
)

type JarCookie cookiejar.Entry

func NewJarCookie(cookie entity.Cookie) JarCookie {
	jarCookie := JarCookie{
		Name:       cookie.Name,
		Value:      cookie.Value,
		Domain:     cookie.Domain,
		Path:       cookie.Path,
		SameSite:   cookie.SameSite,
		Secure:     cookie.Secure,
		HttpOnly:   cookie.HttpOnly,
		HostOnly:   cookie.HostOnly,
		Creation:   cookie.Created,
		LastAccess: cookie.LastAccess,
	}
	if !cookie.IsSession() {
		jarCookie.Expires, jarCookie.Persistent = *cookie.Expires, true
	}
	return jarCookie
}

func NewJarCookies(entries []cookiejar.Entry) []entity.Cookie {
	cookies := make([]entity.Cookie, 0, len(entries))
	for _, entry := range entries {
		cookies = append(cookies, JarCookie(entry).ToEntity())
	}
	return cookies
}

func (j JarCookie) ToEntity() entity.Cookie {
	cookie := entity.Cookie{
		Name:       j.Name,
		Value:      j.Value,
		Domain:     j.Domain,
		Path:       j.Path,
		SameSite:   j.SameSite,
		Secure:     j.Secure,
		HttpOnly:   j.HttpOnly,
		HostOnly:   j.HostOnly,
		Created:    j.Creation,
		LastAccess: j.LastAccess,
	}
	if j.Persistent {
		expires := j.Expires.UTC().Truncate(time.Second)
		cookie.Expires = &expires
	}
	return cookie
}

func (j JarCookie) Entry() cookiejar.Entry {
	return cookiejar.Entry(j)
}
