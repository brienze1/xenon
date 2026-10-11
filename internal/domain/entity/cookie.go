package entity

import "time"

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

func (c Cookie) LogValue() any {
	return map[string]any{"name": c.Name, "domain": c.Domain, "path": c.Path}
}
