package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type Cookie struct {
	Expires    *time.Time `json:"expires,omitempty"`
	Created    time.Time  `json:"created"`
	LastAccess time.Time  `json:"last_access"`
	Name       string     `json:"name"`
	Value      string     `json:"value"`
	Domain     string     `json:"domain"`
	Path       string     `json:"path"`
	SameSite   string     `json:"same_site,omitempty"`
	HostOnly   bool       `json:"host_only"`
	Secure     bool       `json:"secure"`
	HttpOnly   bool       `json:"http_only"`
}

func NewCookies(cookies []entity.Cookie) []Cookie {
	stored := make([]Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		stored = append(stored, Cookie{
			Expires:    cookie.Expires,
			Created:    cookie.Created,
			LastAccess: cookie.LastAccess,
			Name:       cookie.Name,
			Value:      cookie.Value,
			Domain:     cookie.Domain,
			Path:       cookie.Path,
			SameSite:   cookie.SameSite,
			HostOnly:   cookie.HostOnly,
			Secure:     cookie.Secure,
			HttpOnly:   cookie.HttpOnly,
		})
	}
	return stored
}

func (c Cookie) ToEntity() entity.Cookie {
	return entity.Cookie{
		Expires:    c.Expires,
		Created:    c.Created,
		LastAccess: c.LastAccess,
		Name:       c.Name,
		Value:      c.Value,
		Domain:     c.Domain,
		Path:       c.Path,
		SameSite:   c.SameSite,
		HostOnly:   c.HostOnly,
		Secure:     c.Secure,
		HttpOnly:   c.HttpOnly,
	}
}
